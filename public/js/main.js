openpgp.initWorker({ path: 'node_modules/openpgp/dist/openpgp.worker.min.js' })

Vue.use(VueMaterial)

const errUnauthorized = new Error('Unauthorized')
const errCancelled = new Error('Cancelled')

function checkResponse(res) {
	if (res.ok) {
		return res
	}
	if (res.status === 401) {
		throw errUnauthorized
	}
	throw new Error(res.statusText)
}

new Vue({
	el: '#app',
	data: {
		credentials: {
			username: '',
			password: '',
		},
		query: '',
		items: [],
		selectedItem: null,
		keys: null,
		error: null,
	},
	computed: {
		filteredItems() {
			if (this.query === '') {
				return this.items
			}
			return this.items.filter(item => item.indexOf(this.query) !== -1)
		},
	},
	methods: {
		request(url) {
			const h = new Headers()
			if (this.credentials.username !== "" || this.credentials.password !== "") {
				h.set('Authorization', 'Basic '+btoa(this.credentials.username+':'+this.credentials.password))
			}

			return fetch(url, {
				credentials: 'include',
				headers: h,
			})
		},
		show(name) {
			return this.request('pass/store/'+name+'.gpg')
			.then(checkResponse)
			.then(res => res.arrayBuffer())
			.then(buf => this.decrypt(buf))
			.then(text => {
				const metadata = text.trim().split('\n')
				const password = metadata.shift()
				return { name, password, metadata }
			})
			.then(data => this.selectedItem = data)
			.catch(err => {
				if (err != errCancelled) {
					this.showError(err)
				}
			})
		},
		fetchKeys() {
			if (this.keys !== null) {
				return this.keys
			}

			return this.request('pass/keys.gpg')
			.then(checkResponse)
			.then(res => res.arrayBuffer())
			.then(buf => openpgp.armor.encode(openpgp.enums.armor.private_key, new Uint8Array(buf)))
			.then(armored => {
				let { keys, err=[] } = openpgp.key.readArmored(armored)
				if (err.length > 0) {
					throw err[0]
				}

				keys = keys.filter(key => key.verifyPrimaryKey() === openpgp.enums.keyStatus.valid)

				return this.$refs['pgp-ask-pass'].ask()
				.catch(() => {
					throw errCancelled
				})
				.then(passphrase => {
					for (let i = 0; i < keys.length; i++) {
						const key = keys[i]
						if (key.primaryKey.isDecrypted) {
							continue
						}

						if (!key.decrypt(passphrase)) {
							throw new Error('Invalid passphrase')
						}
					}

					this.keys = Promise.resolve(keys)
					return keys
				})
			})
		},
		decrypt(buf) {
			return this.fetchKeys()
			.then(keys => openpgp.decrypt({
				message: openpgp.message.read(new Uint8Array(buf)),
				publicKeys: keys,
				privateKey: keys[0],
			}))
			.then(plaintext => plaintext.data)
		},
		showError(err) {
			console.error(err)
			this.error = err.toString()
			this.$refs['error-bar'].open()
		},
	},
	mounted() {
		const list = () => {
			return this.request('pass/store')
			.then(checkResponse)
			.then(res => res.json())
			.then(items => {
				this.items = items.map(item => item.replace(/\.gpg$/, ''))
			})
			.catch(err => {
				if (err == errUnauthorized) {
					return this.$refs['ask-pass'].ask()
					.catch(() => {
						throw errCancelled
					})
					.then(password => {
						this.credentials.password = password
					})
					.then(list)
				}
				throw err
			})
		}

		return list()
		.catch(this.showError)
	},
})

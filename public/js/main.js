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

function readFile(file) {
	return new Promise((resolve, reject) => {
		const reader = new FileReader()

		reader.addEventListener('load', event => {
			resolve(event.target.result)
		})
		reader.addEventListener('error', event => {
			reject(event.target.error)
		})

		reader.readAsArrayBuffer(file)
	})
}

new Vue({
	el: '#app',
	data: {
		key: '',
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
		request(url, init) {
			if (!init) {
				init = {}
			}
			if (!init.headers) {
				init.headers = new Headers()
			}

			init.credentials = 'include'
			if (this.key !== "") {
				init.headers.set('Authorization', 'Bearer '+this.key)
			}

			return fetch('pass'+url, init)
		},
		auth(username, password) {
			return this.request('/auth', {
				method: 'POST',
				body: JSON.stringify({username, password}),
				headers: new Headers({'Content-Type': 'application/json'}),
			})
			.then(checkResponse)
			.then(res => res.json())
			.then(data => {
				this.key = data.key
			})
		},
		list() {
			return this.request('/store')
			.then(checkResponse)
			.then(res => res.json())
			.then(items => {
				this.items = items.map(item => item.replace(/\.gpg$/, ''))
			})
			.catch(this.showError)
		},
		show(name) {
			return this.request('/store/'+name+'.gpg')
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

			return this.request('/keys.gpg')
			.then(res => {
				if (res.status == 404) {
					return this.$refs['pgp-ask-key'].ask()
					.then(readFile)
				}

				return checkResponse(res)
				.then(res => res.arrayBuffer())
			})
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
		copySelectedPassword() {
			const sel = window.getSelection()

			let ok = false
			try {
				sel.selectAllChildren(this.$refs['password-text'])
				ok = document.execCommand('copy')
			} catch (err) {}
			if (!ok) {
				this.showError('Could not copy password')
			}

			sel.removeAllRanges()
		},
	},
	mounted() {
		this.$nextTick(() => {
			this.$refs['login-ask-pass'].ask()
			.catch(() => {
				throw errCancelled
			})
			.then(creds => this.auth(creds.username, creds.password))
			.then(this.list)
			.catch(this.showError)
		})
	},
})

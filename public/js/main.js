openpgp.initWorker({ path: 'node_modules/openpgp/dist/openpgp.worker.min.js' })

new Vue({
	el: '#app',
	data: {
		query: '',
		items: [],
		selectedItem: null,
		keys: null,
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
		list() {
			return fetch('pass/store')
			.then(res => res.json())
			.then(list => {
				this.items = list.map(item => item.replace(/\.gpg$/, ''))
			})
		},
		show(name) {
			return fetch('pass/store/'+name+'.gpg')
			.then(res => res.arrayBuffer())
			.then(buf => this.decrypt(buf))
			.then(text => {
				const metadata = text.trim().split('\n')
				const password = metadata.shift()
				return { name, password, metadata }
			})
			.then(data => this.selectedItem = data)
		},
		fetchKeys() {
			if (this.keys !== null) {
				return Promise.resolve(this.keys)
			}

			return fetch('pass/keys.gpg')
			.then(res => res.arrayBuffer())
			.then(buf => openpgp.armor.encode(openpgp.enums.armor.private_key, new Uint8Array(buf)))
			.then(armored => {
				let { keys, err=[] } = openpgp.key.readArmored(armored)
				if (err.length > 0) {
					throw err[0]
				}

				keys = keys
				.filter(key => key.verifyPrimaryKey() === openpgp.enums.keyStatus.valid)
				.map(key => {
					while (!key.primaryKey.isDecrypted) {
						// TODO: use a password entry
						let passphrase = prompt('Password')
						if (!passphrase) {
							throw new Error('Cancelled')
						}
						if (key.decrypt(passphrase)) {
							break
						}
					}
					return key
				})

				this.keys = keys
				return keys
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
	},
	created() {
		this.list()
	},
})

Vue.component('ask-pass', {
	template: '#ask-pass-template',
	data: () => {
		return {
			passphrase: '',
			submit: () => {},
			cancel: () => {},
		}
	},
	props: ['title', 'description'],
	methods: {
		ask() {
			const dialog = this.$refs.dialog

			return new Promise((resolve, reject) => {
				this.submit = () => {
					const passphrase = this.passphrase
					dialog.close()
					resolve(passphrase)
				}

				this.cancel = () => {
					dialog.close()
					reject()
				}

				dialog.open()
			})
		},
	},
	mounted() {
		const dialog = this.$refs.dialog

		dialog.$on('open', () => {
			this.$nextTick(() => this.$refs.passphrase.$el.focus())
		})

		dialog.$on('close', () => {
			this.passphrase = ''
		})
	},
})

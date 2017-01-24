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
			const dialog = this.$refs['ask-pass-dialog']

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
		const dialog = this.$refs['ask-pass-dialog']

		dialog.$on('open', () => {
			setTimeout(() => {
				this.$refs['passphrase'].$el.focus()
			}, 0)
		})

		dialog.$on('close', () => {
			this.passphrase = ''
		})
	},
})

Vue.component('ask-creds', {
	template: '#ask-creds-template',
	data: () => {
		return {
			username: '',
			password: '',
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
					const creds = {
						username: this.username,
						password: this.password,
					}
					dialog.close()
					resolve(creds)
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
			this.$nextTick(() => this.$refs.username.$el.focus())
		})

		dialog.$on('close', () => {
			this.username = ''
			this.password = ''
		})
	},
})

Vue.component('ask-file', {
	template: '#ask-file-template',
	data: () => {
		return {
			submit: () => {},
			cancel: () => {},
		}
	},
	props: ['title', 'description'],
	methods: {
		ask() {
			const dialog = this.$refs['ask-file-dialog']

			return new Promise((resolve, reject) => {
				this.submit = (event) => {
					const file = event.target.files[0]
					if (file) {
						resolve(file)
						dialog.close()
					}
				}

				this.cancel = () => {
					dialog.close()
					reject()
				}

				dialog.open()
			})
		},
	},
})

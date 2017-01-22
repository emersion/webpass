Vue.component('ask-pass', {
	template: '#ask-pass-template',
	data: () => {
		return {
			passphrase: '',
		}
	},
	methods: {
		submit() {
			this.$emit('submit', this.passphrase)
		},
		cancel() {
			this.$emit('submit', null)
		},
	},
})

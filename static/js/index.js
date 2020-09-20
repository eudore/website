//


Vue.component("Container", {
	template: `
	<main id="container">
		<div id="index-welcome">
			<span>Welcome to the eudore website!</span>
			<span>eudore website is a multifunction platform.</span>
		</div>
		<div id="index-overview">
		</div>
	</main>
	`,
})
var vm = new Vue({
	el: '#app',
	template: `<App></App>`,
	data: function() {
		return {
			Base: Base,
		}
	},
})
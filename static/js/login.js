//

"use strict";
var apiVersion = "/api/v1/auth"

function geRedirect() {
	var redirect = new URL(window.location.href).searchParams.get('location') + window.location.hash || ''
	if (window.location == redirect) {
		redirect = ""
	}
	if (Base.bearer != "" && redirect != "") {
		window.location = redirect
	}
	return redirect
}

Vue.component("Container", {
	data: function() {
		return {
			username: "",
			password: "",
			captcha: "",
			verify: "",
			captchaData: null,
			redirect: geRedirect(),
		}
	},
	mounted: function() {
		this.reloadCaptcha()
	},
	computed: {
		getLocation: function() {
			if (this.redirect) {
				return "?location=" + this.redirect
			}
			return ""
		}
	},
	methods: {
		login: function() {
			axios.post(apiVersion + "/login/website", {
				username: this.username,
				password: this.password,
				captcha: this.captcha,
				verify: this.verify,
			}).then((response) => {
				console.log(response)
				if (response.data.bearer.length > 10) {
					localStorage.setItem("user", JSON.stringify(response.data))
					window.location = this.redirect
				}
			})
		},
		reloadCaptcha: function() {
			axios.get(apiVersion + "/login/captcha", {
				headers: {
					Accept: "application/base64"
				},
			}).then((response) => {
				this.verify = response.headers.captcha
				this.captchaData = response.data
			})
		},
		oauth2: function(addr) {
			event.preventDefault();
			window.open(addr, 'new', 'width=500,height=600,menubar=no,toolbar=no,location=no,directories=no,status=no,scrollbars=yes,resizable=yes');
		}
	},
	template: `
<main id="container">
	<form id="slick-login">
		<label for="name">用户名</label>
		<input type="text" v-model=username placeholder="username" name="username" class="placeholder">
		<label for="pass">password</label>
		<input type="password" v-model=password placeholder="password" name="password" class="placeholder">
		<label for="captcha"></label>
		<div>
			<input type="text" v-model=captcha placeholder="captcha" name="captcha" class="placeholder">
			<img @click="reloadCaptcha" :src="captchaData">
		</div>
		<input @click="login" type="button" value="Sign in">
		<div>
			<a @click="oauth2('/api/v1/auth/login/oauth2/github')" :href="'/api/v1/auth/login/oauth2/github' + getLocation">
				<svg width="36" height="36" viewBox="0 0 16 16" version="1.1">
					<path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8z"></path>
				</svg>
			</a>
			<a @click="oauth2('/api/v1/auth/login/oauth2/gitlab')" :href="'/api/v1/auth/login/oauth2/gitlab' + getLocation">
				<svg width="36" height="36" viewBox="0 0 36 36">
					<path fill="#e24329" d="M2 14l9.38 9v-9l-4-12.28c-.205-.632-1.176-.632-1.38 0z"></path>
					<path fill="#e24329" d="M34 14l-9.38 9v-9l4-12.28c.205-.632 1.176-.632 1.38 0z"></path>
					<path fill="#e24329" d="M18,34.38 3,14 33,14 Z"></path>
					<path fill="#fc6d26" d="M18,34.38 11.38,14 2,14 6,25Z"></path>
					<path fill="#fc6d26" d="M18,34.38 24.62,14 34,14 30,25Z"></path>
					<path fill="#fca326" d="M2 14L.1 20.16c-.18.565 0 1.2.5 1.56l17.42 12.66z"></path>
					<path fill="#fca326" d="M34 14l1.9 6.16c.18.565 0 1.2-.5 1.56L18 34.38z"></path>
				</svg>
			</a>
		</div>
	</form>
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
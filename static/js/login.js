"use strict";
var apiVersion = "/api/v1/auth"
var redirect = new URL(window.location.href).searchParams.get('location') || ''
if(window.location == redirect) {
	redirect = ""
}
if(Base.bearer != "" && redirect != "") {
	window.location = redirect
}

var Index = {
	view: function(vnode) {
		return m("form#slick-login", [
			m("label", {for: "name"}, I18n("username")),
			m("input.placeholder", {type: "text", name: "name", placeholder: "username"}),
			m("label", {for: "pass"}, "password"),
			m("input.placeholder", {type: "password", name: "pass", placeholder: "password"}),
			m("label", {for: "captcha"}),
			m("div", [
				m("input.placeholder", {type: "text", name: "captcha", placeholder: "captcha"}),
				m("img", {src: vnode.state.captcha, oninit: function() {
					reloadCaptcha(vnode)
				}, onclick() {
					console.log(vnode)
					reloadCaptcha(vnode)
				}}),
			]),
			m("input", {type: "button", value: "Sign in", onclick: function(){
				m.request({
					method: "POST",
					url: apiVersion + "/login/wejass",
					data: {
						name: input[1].value,
						pass: input[2].value,
						captcha: input[3].value,
						verify: vnode.state.verify
					},
				}).then(function(result){
					console.log(result)
					if(result.bearer.length > 10) {
						localStorage.setItem("user", JSON.stringify(result))
						window.location = redirect
					}
				})
			}}),
			m("div", [
				m("a", {href: "/auth/login/github" + (redirect=="" ? "": ("?location=" +redirect))}, m("svg", {
					width: 36,
					height: 36,
					viewBox: "0 0 16 16",
					version: "1.1",
					// "aria-hidden": true
				}, m("path", {
					// "fill-rule": "evenodd",
					d: "M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8z",
				})),
				),
				m("svg", {
					width: 36,
					height: 36,
					viewBox: "0 0 36 36",
				}, [
					m("path", {fill: "#e24329", d: "M2 14l9.38 9v-9l-4-12.28c-.205-.632-1.176-.632-1.38 0z"}),
					m("path", {fill: "#e24329", d: "M34 14l-9.38 9v-9l4-12.28c.205-.632 1.176-.632 1.38 0z"}),
					m("path", {fill: "#e24329", d: "M18,34.38 3,14 33,14 Z"}),
					m("path", {fill: "#fc6d26", d: "M18,34.38 11.38,14 2,14 6,25Z"}),
					m("path", {fill: "#fc6d26", d: "M18,34.38 24.62,14 34,14 30,25Z"}),
					m("path", {fill: "#fca326", d: "M2 14L.1 20.16c-.18.565 0 1.2.5 1.56l17.42 12.66z"}),
					m("path", {fill: "#fca326", d: "M34 14l1.9 6.16c.18.565 0 1.2-.5 1.56L18 34.38z"}),
				]),
			])
		])
	}
}

function reloadCaptcha(vnode) {
	m.request({
		method: "GET",
		url: apiVersion + "/captcha",
		headers: {
			Accept: "application/base64",
		},
		extract: allxhr,
	}).then(function(result) {
		vnode.state.verify = result.header.captcha
		vnode.state.captcha = result.body 
	})
}

var input
window.onload = function() {
	m.mount(document.body, Home)
	m.mount(document.getElementById('container'), Index)
	input=document.getElementsByTagName('input')
	document.onkeydown=function(e){
		var keyCode = e.keyCode || e.which || e.charCode;
		var ctrlKey = e.ctrlKey || e.metaKey;
		if(ctrlKey && keyCode == 67) {
			var code = input[2]
			if(code.style.display=='none')
				code.style.display='block'
			else
				code.style.display='none'
			e.preventDefault();
		}
	}; 
}
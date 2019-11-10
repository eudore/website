"use strict";
var apiVersion = "/api/v1/auth"

var Nav = {
	nav: ["user", "oauth2"],
	view: function() {
		return [
			m("div#setting-nav", m("ul", Nav.nav.map(function(i) {
				return m("li", m("a", {href: "#!/" + i}, i))
			}))),
			m("div#setting-content")
		]
	}
}

var Setting = {
	data: {},
	table: new TableData({}),
	User: {
		oninit: function(){
			Setting.getSetting()

			Setting.table.keys = ["name", "icon", "mail", "lang", "time"]
			Setting.table.vals = ["name", "icon", "mail", "lang", "time"]
			Setting.table.setpack("name", function(i) {
				return m("a", {href: ""}, i)
			})
			Setting.table.setpack("lang", function(i) {
				return i || "en-US"
			})
			Setting.table.setpack("icon", function() {
				return m("div", [
					m("img.avatar", {src: "/api/v1/auth/user/icon/name/" + Base.name})
				])
			})
			// Setting.table.redraw()
		},
		view: function() {
			return [
				m("div", "ss"),
				m("div.column", [
					m(Setting.table),
				]),
			]
		}
	},
	Oauth2: {
		oninit: function() {
			Setting.getOauth2()
		},
		view: function() {
			return [
				m("div", "login"),
				m("div.column", [
					m("div.row", [
						m("span", "github"),
						m("span", "绑定"),
					]),
					m("div.row", [
						m("span", "google"),
						m("span", "绑定"),
					]),
				])
			]
		}
	},
	getSetting: function() {
		m.request({
			method: "GET",
			url: apiVersion + "/user/setting",
		}).then(function(result){
			Setting.table.data = result
		})
	},
	getOauth2: function() {
		m.request({
			method: "GET",
			url: apiVersion + "/oauth2/bind"
		})
	}
}
window.onload = function() {
	m.mount(document.body, Home)
	m.mount(document.getElementById('container'), Nav)
	m.route(document.getElementById('setting-content'), "/user", {
		"/user": Setting.User,
		"/oauth2": Setting.Oauth2
	})
}

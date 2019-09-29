"use strict";
var apiVersion = "/api/v1/status"
var Build = {
	oninit: function(vnode) {
		vnode.state.data = []
		this.fetch(vnode)
	},
	fetch: function(vnode) {
		m.request({
			method: "GET", 
			url: apiVersion + "/build"
		}).then(function(result){
			vnode.state.data = result
		})
	},
	view: function(vnode) {
		return m("div#status-build", [
			m("fieldset", [
				m("legend", I18n("Build Info")),
				m("div", vnode.state.data.map(function(i){
					return m("tr", [
						m("td", I18n(i.name)),
						m("td", m("a", {href: i.link}, i.version))
					])
				}))
			])
		])
	}
}
var System = {
	keys: ['Uptime', 'NumGoroutine', 'MemAllocated', 'MemTotal', 'MemSys', 'Lookups', 'MemMallocs', 'MemFrees', 'HeapAlloc', 'HeapSys', 'HeapIdle', 'HeapInuse', 'HeapReleased', 'HeapObjects','StackInuse', 'StackSys', 'MSpanInuse', 'MSpanSys', 'MCacheInuse', 'MCacheSys', 'BuckHashSys', 'GCSys', 'OtherSys', 'NextGC', 'LastGC', 'PauseTotalNs', 'PauseNs', 'NumGC'],
	oninit: function(vnode) {
		vnode.state.data = []
		this.fetch(vnode)
	},
	fetch: function(vnode) {
		m.request({
			method: "GET", 
			url: apiVersion + "/system"
		}).then(function(result){
			vnode.state.data = result
		})
	},
	view: function(vnode) {
		return m("div#status-system", [
			m("fieldset", [
				m("legend", I18n("System Info")),
				vnode.state.data ? System.keys.map(function(i){
					return m("tr", [
						m("td", I18n(i) ),
						m("td", vnode.state.data[i])
					])
				}) : ""
			])
		])
	}
}
var Config = {
	oninit: function(vnode) {
		vnode.state.data = []
		this.fetch(vnode)
	},
	fetch: function(vnode) {
		m.request({
			method: "GET", 
			url: apiVersion + "/config"
		}).then(function(result){
			vnode.state.data = result
		})
	},
	view: function(vnode) {
		return m("div#status-config", [
			m("fieldset", [
				m("legend", I18n("Config Info")),
				m("div", vnode.state.data.map(function(i){
					return m("tr", [
						m("td", i.name),
						m("td", i.value)
					])
				}))
			])
		])
	}

}

var Index = {
	view: function(vnode) {
		return [
			m(Build),
			m(System),
			m(Config)
		]
	}
}

m.mount(document.body, Home)
m.route(document.getElementById('container'), "/index", {
	"/index": Index
})
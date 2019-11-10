"use strict";
var apiVersion = "/api/v1/task"

var Scheduler = {
	view: function() {
		return m("div", "Trigger")
	}
}
var Trigger = {
	// data: {id:0, level:0, name:"", mail:"",  loginip:0, logintime:0},
	data: new TableData({}),
	table: new Tables("/trigger/index", "/trigger/count"),
	Index: {
		oninit: function(vnode) {
			Trigger.table.head = ["id", "name", "event", "description", "params", "schedule", "executorid", "time"]
			Trigger.table.line = ["id", "name", "event", "description", "params", "schedule", "executorid", "time"]
			Trigger.table.setpack("name", function(i) {
				return m("a", {href: "#!/trigger/info/" + i}, i)
			})
			Trigger.table.setpack("status", function(i) {
				return statusstr(i)
			})
			Trigger.table.setpack("logintime", function(i) {
				return new Date(i).Format("A dd,yyyy")
			})
			Trigger.table.redraw()
		},
		view: function(vnode) {
			return m("div#trigger-index", [
				m("div", [
					m("input#trigger-search"),
					m("button", "search key"),
				]),
				m(Trigger.table)
			])
		}
	},
	Info: {
		oninit: function(vnode){
			Trigger.data.keys = ["id", "name", "description", "event", "params", "schedule", "executorid", "time"]
			Trigger.data.vals = ["id", "name", "description", "event", "params", "schedule", "executorid", "time"]
			Trigger.data.setpack("params", function(i) {
				return m("pre", JSON.stringify(JSON.parse(i || '{}'), null, 4))
			})
			Trigger.data.setpack("schedule", function(i) {
				return i ? i : "all"
			})
			Trigger.getTrigger(vnode.attrs.name)
		},
		view: function(){
			return m("div", m(Trigger.data))
		}
	},
	// controller
	getTrigger: function(name) {
		m.request({
			method: "GET", 
			url: apiVersion + "/trigger/info/name/" + name,
		}).then(function(result){
			Trigger.data.data = result
		})
	},
}
var Executor = {
	data: new TableData({}),
	table: new Tables("/executor/index", "/executor/count"),
	Index: {
		oninit: function(vnode) {
			Executor.table.head = ["id", "name", "type", "description", "config", "time"]
			Executor.table.line = ["id", "name", "type", "description", "config", "time"]
			Executor.table.setpack("name", function(i) {
				return m("a", {href: "#!/executor/info/" + i}, i)
			})
			Executor.table.setpack("time", function(i) {
				return new Date(i).Format("A dd,yyyy")
			})
			Executor.table.redraw()
		},
		view: function(vnode) {
			return m("div#executor-index", [
				m("div", [
					m("input#executor-search"),
					m("button", "search key"),
				]),
				m(Executor.table)
			])
		}
	},
	Info: {
		oninit: function(vnode){
			Executor.data.keys = ["id", "name", "description", "type", "config", "time"]
			Executor.data.vals = ["id", "name", "description", "type", "config", "time"]
			Executor.data.setpack("config", function(i) {
				return m("pre", JSON.stringify(JSON.parse(i || '{}'), null, 4))
			})
			Executor.getExecutor(vnode.attrs.name)
		},
		view: function(){
			return m("div", m(Executor.data))
		}
	},
	Exec: {
		oninit: function(vnode) {

			Executor.getExecutor(vnode.attrs.name)
		},
		view: function(vnode) {
			return m("div#task-execytor-exec", [
				m("div", vnode.attrs.name),
				m("textarea"),
				m("pre"),
				m("button", {onclick: Executor.putExecMessage}, "send")
			])
		}
	},
	// controller
	getExecutor: function(name) {
		m.request({
			method: "GET", 
			url: apiVersion + "/executor/info/name/" + name,
		}).then(function(result){
			Executor.data.data = result
		})
	},
	putExecMessage: function(name) {
		m.request({
			method: "PUT",
			url: apiVersion + '/executor/exec/' + Executor.data.data.name
			
		})
	}
}

var Logger = {
	table: new Tables("/task/index", "/task/count"),
	status: ["成功", "失败", "取消"],
	oninit: function(vnode) {
		Logger.table.head = ["status", "params", "starttime", "endtime", "eventid", "executorid", "message"]
		Logger.table.line = ["status", "params", "starttime", "endtime", "eventid", "executorid", "message"]
		Logger.table.setpack("status", function(i) {
			return Logger.status[i]
		})
		Logger.table.setpack("starttime", function(i) {
			return new Date(i).Format("yyyy-MM-dd hh:mm:ss")
		})
		Logger.table.setpack("endtime", function(i) {
			return new Date(i).Format("yyyy-MM-dd hh:mm:ss")
		})
		Logger.table.setpack("message", function(i) {
			return m("div", {data: i}, m("button", "show"))
		})
		Logger.table.redraw()

		// Logger.getLogerList()
	},
	view: function() {
		return m(Logger.table)
	},
	// controller
	getLogerList: function() {
		m.request({
			method: "GET",
			url: apiVersion + "/task/index",
		}).then(function(result){
			console.log(result)
		})
	}
}

var Index = {
	view: function(vnode) {
		return [
		]
	}
}

var Nav = {
	menu: ["Overview", "Trigger", "Scheduler", "Executor", "Logger"],
	view: function() {
		return [
			m("nav#task-nav.box", Nav.menu.map(function(i){
				return m("a", {href: "#!/" + i.toLowerCase()}, i)
			})),
			m("div#task-content.box")
		]
	}
}

window.onload = function() {
	m.mount(document.body, Home)
	m.mount(document.getElementById('container'), Nav)
	m.route(document.getElementById('task-content'), "/index", {
		"/index": Index,
		"/trigger": Trigger.Index,
		"/trigger/info/:name": Trigger.Info,
		"/executor": Executor.Index,
		"/executor/info/:name": Executor.Info,
		"/executor/exec/:name": Executor.Exec,
		"/logger": Logger,
	})
}
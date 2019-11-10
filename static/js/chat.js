"use strict";
var apiVersion = "/api/v1/chat"


var Index = {
	Select: 0,
	UserList: [],
	UserName: {
		1: "eudore",
		2: "root",
		4: "user02",
		101: "guest",
	},
	LastTime: {},
	LastMessage: {},
	Messages: {},
	Connect: null,
	oninit: function() {
		Index.getMessageList()
		Index.getConnect()
	},
	view: function() {
		return m("div#chat-container", [
			m("div#chat-userlist", Index.UserList.map(function(i){
				return m(Index.Select==i?"li.chat-userlist-select":"li", {onclick: function(){Index.Select = i}}, [
					m("img.avatar", {src: "/api/v1/auth/user/icon/id/" + i}),
					m("div.chat-userlist-info", [
						m("div.chat-userlist-top", [
							m("span.chat-userlist-name", Index.UserName[i] || i),
							m("span.chat-userlist-time", Index.timeFormat(Index.LastTime[i]))
						]),
						m("div.chat-userlist-message", Index.LastMessage[i])
					])
				])
			})),
			m("div#chat-message", [
				m("div#chat-message-input", [
					m("textarea#chat-message-edit"),
					m("button.button#chat-message-button", {onclick: Index.sendMessage}, "send"),
				]),
				m("div#chat-message-list", (Index.Messages[Index.Select] || []).map(function(i){
					return m("li" + (i.sendid != i.id ? ".chat-message-send" : ""), [
						m("img.avatar", {src: "/api/v1/auth/user/icon/id/" + i.id}),
						m("div.column", [
							m("span", Index.UserName[i.id] || i.id),
							m("span", i.message),
						]),
						// m("span", Index.timeFormat(i.time))
					])
				})),
			])
		])
	},


	getConnect: function(code){
		if(Index.Connect==null){ 
			var wsUrl = location.host + apiVersion + "/message/connect?bearer=" + Base.bearer;
			if(location.protocol == "http:") {
				wsUrl = "ws://" + wsUrl
			}else {
				wsUrl = "wss://" + wsUrl
			}
			Index.Connect = new WebSocket(wsUrl);
			try {
				Index.Connect.onopen = function () {
					console.log("服务器-连接")
					if(code){
						code()
					}
				}
				Index.Connect.onmessage = function (event) {
					var data = JSON.parse(event.data);
					Index.addMessage(data)
					m.redraw()
				}
				Index.Connect.onclose = function () {
					if (Index.Connect) {
						Index.Connect.close();
						Index.Connect = null;
					}
					console.log("服务器-关闭")
				}
				Index.Connect.onerror = function () {
					if (Index.Connect) {
						Index.Connect.close();
						Index.Connect = null;
					}
					console.log("服务器-错误")
				}
			} catch (e) {
				alert(e.message);
			}
		}else if(code){
			code()
		}
	},

	//
	sendMessage: function() {
		var msg = document.getElementById("chat-message-edit").value
		if(msg=="" || Index.Select == 0) {
			return
		}
		var selectuser = Index.Select
		Index.getConnect(function(){
			var request = JSON.stringify({receid: selectuser, message: msg})
			Index.Connect.send(request)
		})
	},
	timeFormat: function(str) {
		var current = new Date
		var obj = new Date(str)
		if(current.getFullYear() == obj.getFullYear()) {
			if(current.getMonth() == obj.getMonth() && current.getDate() == obj.getDate()) {
				return obj.Format("hh:mm")
			}
			return obj.Format("MM-dd")
		}
		return obj.Format("yyyy-MM-dd")
	},
	getMessageList: function() {
		m.request({
			method: "GET",
			url: apiVersion + "/message/list",
		}).then(function(result){
			for(var i of result) {
				Index.addMessage(i)
			}
		})
	},
	addMessage: function(data) {
		var id = data.sendid
		if(data.sendid == Base.userid) {
			id = data.receid
		}
		data.id = id
		if(Index.UserList.indexOf(id)==-1){
			Index.UserList.push(id)
			Index.Messages[id] = []
		}
		Index.LastTime[id] = data.time
		Index.LastMessage[id] = data.message.slice(0, 24)
		Index.Messages[id].push(data)

	}
}

window.onload = function() {
	m.mount(document.body, Home)
	m.mount(document.getElementById('container'), Index)
}
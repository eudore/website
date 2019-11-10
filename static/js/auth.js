"use strict";
var apiVersion = "/api/v1/auth"
var statuslist = ["正常", "冻结", "禁用"]
function statusstr(status) {
	return statuslist[status]
}
var User = {
	// model
	data: {id:0, level:0, name:"", mail:"",  loginip:0, logintime:0},
	table: new Tables("/user/index", "/user/count"),
	bind: {
		permission: [],
		role: [],
		policy: [],
	},
	grant: {
		permission: [],
		role: [],
		policy: [],
	},
	// view
	Index: {
		oninit: function(vnode) {
			User.table.head = ["id", "name", "status", "mail", "loginip", "logintime"]
			User.table.line = ["id", "name", "status", "mail", "loginip", "logintime"]
			User.table.setpack("name", function(i) {
				return m("a", {href: "#!/user/info/" + i}, i)
			})
			User.table.setpack("status", function(i) {
				return statusstr(i)
			})
			User.table.setpack("logintime", function(i) {
				return new Date(i).Format("A dd,yyyy")
			})
			User.table.redraw()
		},
		view: function(vnode) {
			return m("div", [
				m("div", [
					m("input#user-search"),
					m("button", {onclick: User.searchPermission},"search key"),
				]),
				m(User.table)
			])
		}
	},
	Info: {
		oninit: function(vnode) {
			User.getUser(vnode.attrs.name)
			User.getUserPermission(vnode.attrs.name)
			User.getUserRole(vnode.attrs.name)
			User.getUserPolicy(vnode.attrs.name)
		},
		view: function(vnode) {
			return m("div#user-info", [
				m("fieldset", [
					m("legend", "basic info"),
					m("div", [
						 User.data["name"] ? m("img", {src: apiVersion + "/user/icon/name/" +  User.data["name"], width: 40}) : "",
						m("a", {href: "#!/user/edit/" + User.data["name"]}, "修改"),
						m("p", JSON.stringify(User.data)),
						m("p", "用户登录方式"),
						m("p", "用户权限"),
						m("ul.well-list", [
							m("li", [m("span.light", "Name:"), m("strong", User.data["name"])]),
							m("li", [m("span", "Mail:"), m("strong", User.data["mail"])]),
							m("li", [m("span", "Login ip:"), m("strong", User.data["loginip"])]),
							m("li", [m("span", "Login time:"), m("strong", User.data["logintime"])])
						]),
					])
				]),
				
				m("fieldset", [
					m("legend", "Permission"),
					m("div", [
						m("a", {href: "#!/user/grant/permission/" + vnode.attrs.name}, "grant"),
						User.bind.permission ? User.bind.permission.map(function(i){
							return m("tr", [
								m("td", m("a", {href: "#!/permission/info/" + i.name}, i.name)),
								m("td", i.effect),
								m("td", i.description),
								m("td", m("button.button", {onclick: function(){User.removePermission(i.id)}},"remove"))
							])
						}) : ""
					])
				]),

				m("fieldset", [
					m("legend", "Role"),
					m("div", [
						m("a", {href: "#!/user/grant/role/" + vnode.attrs.name}, "grant"),
						User.bind.role ? User.bind.role.map(function(i){
							return m("tr", [
								m("td", m("a", {href: "#!/role/info/" + i.name}, i.name)),
								m("td", i.description),
								m("td", m("button.button", {onclick: function(){User.removeRole(i.id)}},"remove"))
							])
						}) : ""
					])
				]),

				m("fieldset", [
					m("legend", "Policy"),
					m("div", [
						m("a", {href: "#!/user/grant/policy/" + vnode.attrs.name}, "grant"),
						User.bind.policy ? User.bind.policy.map(function(i){
							return m("tr", [
								m("td", m("a", {href: "#!/policy/info/" + i.name}, i.name)),
								m("td", i.description),
								m("td", m("button.button", {onclick: function(){User.removePolicy(i.id)}},"remove"))
							])
						}) : ""
					])
				])
			])
		}
	},
	Edit: {
		oninit: function(vnode) {
			m.request({
				method: "GET",
				url: apiVersion + "/user/info/name/" + vnode.attrs.name,
				// extract: allxhr
			}).then(function(result){
				console.log(result)
				vnode.state.data = result
			})
		},
		submit: function(){
			console.log(vnode.state.data)
		},
		view: function(vnode) {
			return vnode.state.data ? m("div.user-edit", [
				m("h2.page-title", vnode.state.data["name"]),
				m("div.user-edit", [
					m("fieldset", [
						m("legend", "Account"),
						m("div", [
							m("label.control-label", "Name"),
							m("div", m("input.input", {type: "text", value: vnode.state.data["name"] }))
						]),
						m("div", [
							m("label.control-label", "Name"),
							m("div", m("input.input", {type: "text", value: vnode.state.data["name"]}))
						])
					]),
					m("fieldset", [
						m("legend", "Password"),
						m("div.box", [
							m("label.control-label", "Password"),
							m("div", m("input.input", {type: "password"}))
						]),
						m("div.box", [
							m("label.control-label", "Password confirmation"),
							m("div", m("input.input", {type: "password"}))
						])
					]),
					m("button", "Save changes")
				])
			]) : ""
		}
	},
	GrantPermission: {
		oninit: function(vnode) {
			User.data.name = vnode.attrs.name
			User.getUser(vnode.attrs.name)
			User.getUserPermission(vnode.attrs.name)
			User.getListPermission()
		},
		view: function() {
			return m("div#user-grant-permission", [
				m("div.box", [
					m("div", User.grant.permission.map(function(i){
						return !i.fouce ? m("tr", [
							m("td", i.effect),
							m("td", i.name),
							m("td", i.description),
							m("td", m("button.button", {onclick:function(){i.fouce="allow"}}, "allow")),
							m("td", m("button.button", {onclick:function(){i.fouce="deny"}}, "deny")),
						]) : ""
					})),
					m("div#grant-permission", User.grant.permission.map(function(i){
						return i.fouce ? m("tr", [
							m("td", i.effect),
							m("td", i.name),
							m("td", i.description),
							m("td", m("button.button", {onclick:function(){i.fouce=undefined}}, "cancel")),
						]) : ""
					}))
				]),
				m("div", [
					m("button.button", {onclick: User.grantPermission}, "commit"),
					m("button.button", {onclick: function() {window.history.back()} }, "cancel")
				])
			])
		}
	},
	GrantRole: {
		oninit: function(vnode) {
			User.data.name = vnode.attrs.name
			User.getUser(vnode.attrs.name)
			User.getUserRole(vnode.attrs.name)
			User.getListRole()
		},
		view: function() {
			return m("div#user-grant-policy", [
				m("div.box", [
					m("div", User.grant.role.map(function(i){
						return !i.fouce ? m("tr", [
							m("td", i.name),
							m("td", i.description),
							m("td", m("button.button", {onclick:function(){i.fouce=true}}, "grant")),
						]) : ""
					})),
					m("div#grant-role", User.grant.role.map(function(i){
						return i.fouce ? m("tr", [
							m("td", i.name),
							m("td", i.description),
							m("td", m("button.button", {onclick:function(){i.fouce=false}}, "cancel")),
						]) : ""
					}))
				]),
				m("div", [
					m("button.button", {onclick: User.grantRole}, "commit"),
					m("button.button", {onclick: function() {window.history.back()} }, "cancel")
				])
			])
		}
	},
	GrantPolicy: {
		oninit: function(vnode) {
			User.data.name = vnode.attrs.name
			User.getUser(vnode.attrs.name)
			User.getUserPolicy(vnode.attrs.name)
			User.getListPolicy()
		},
		view: function() {
			return m("div#user-grant-policy", [
				m("div.box", [
					m("div", User.grant.policy.map(function(i){
						return !i.fouce ? m("tr", [
							m("td", i.name),
							m("td", i.description),
							m("td", m("button.button", {onclick:function(){i.fouce=true}}, "grant")),
						]) : ""
					})),
					m("div#grant-policy", User.grant.policy.map(function(i){
						return i.fouce ? m("tr", [
							m("td", i.name),
							m("td", i.description),
							m("td", m("button.button", {onclick:function(){i.fouce=false}}, "cancel")),
						]) : ""
					}))
				]),
				m("div", [
					m("button.button", {onclick: User.grantPolicy}, "commit"),
					m("button.button", {onclick: function() {window.history.back()} }, "cancel")
				])
			])
		}
	},
	// controller
	getUser: function(name) {
		m.request({
			method: "GET", 
			url: apiVersion + "/user/info/name/" + name,
		}).then(function(result){
			User.data = result
		})
	},

	// permission
	getUserPermission: function(name) {
		m.request({
			method: "GET",
			url: apiVersion + "/user/permission/name/" + name,
		}).then(function(result){
			User.bind.permission = result
			User.grant.permission = User.grant.permission.filter(function(v) {
				return User.bind.permission.findIndex(function(elem){
					return elem.id == v.id
				}) == -1
			})
		})
	},
	getListPermission: function() {
		m.request({
			method: "GET",
			url: apiVersion + "/permission/list",
		}).then(function(result){
			User.grant.permission = result.filter(function(v) {
				return User.bind.permission.findIndex(function(elem){
					return elem.id == v.id
				}) == -1
			})
		})
	},
	grantPermission: function() {
		var data = []
		for(var i of User.grant.permission) {
			if(typeof i.fouce !== 'undefined') {
				data.push({id: i.id, effect: i.fouce})
			}
		}
		if(data.length == 0) {
			return
		}

		m.request({
			method: "PUT",
			url: apiVersion + "/user/bind/permission/" + User.data.id,
			data: data,
			message: {
				succes: "user "+ User.data.name +" grant permission succes.",
				fatal: "user "+ User.data.name +" grant permission fatal.",
			}
		}).then(function() {
			window.history.back();
		})
	},
	removePermission: function(id) {
		m.request({
			method: "DELETE",
			url: apiVersion + "/user/bind/permission/" + User.data.id + "/" + id,
			message: {
				succes: "remove permission succes",
				fatal: "remove permission fatal",
			}
		}).then(function() {
			User.bind.permission = User.bind.permission.filter(i => i.id!=id)
		})
	},

	// role
	getUserRole: function(name) {
		m.request({
			method: "GET",
			url: apiVersion + "/user/role/name/" + name,
		}).then(function(result){
			User.bind.role = result
		})
	},
	getListRole: function(){
		m.request({
			method: "GET",
			url: apiVersion + "/role/list",
		}).then(function(result){
			User.grant.role = result.filter(function(v) {
				return User.bind.role.findIndex(function(elem){
					return elem.id == v.id
				}) == -1
			})
		})

	},
	grantRole: function(){
		var data = []
		for(var i of User.grant.role) {
			if(i.fouce) {
				data.push({id: i.id})
			}
		}
		if(data.length == 0) {
			return
		}

		m.request({
			method: "PUT",
			url: apiVersion + "/user/bind/role/" + User.data.id,
			data: data,
			message: {
				succes: "user "+ User.data.name +" grant role succes.",
				fatal: "user "+ User.data.name +" grant role fatal.",
			}
		}).then(function() {
			window.history.back();
		})
	},
	removeRole: function(id) {
		m.request({
			method: "DELETE",
			url: apiVersion + "/user/bind/role/" + User.data.id + "/" + id,
			message: {
				succes: "remove role succes",
				fatal: "remove role fatal",
			}
		}).then(function() {
			User.bind.role = User.bind.role.filter(i => i.id!=id)
		})
	},

	// policy
	getUserPolicy: function(name) {
		m.request({
			method: "GET",
			url: apiVersion + "/user/policy/name/" + name,
		}).then(function(result){
			User.bind.policy = result
		})
	},
	getListPolicy: function(){
		m.request({
			method: "GET",
			url: apiVersion + "/policy/list",
		}).then(function(result){
			User.grant.policy = result.filter(function(v) {
				return User.bind.policy.findIndex(function(elem){
					return elem.id == v.id
				}) == -1
			})
		})
	},
	grantPolicy: function() {
		var data = []
		for(var i of User.grant.policy) {
			if(i.fouce) {
				data.push({id: i.id})
			}
		}
		if(data.length == 0) {
			return
		}

		m.request({
			method: "PUT",
			url: apiVersion + "/user/bind/policy/" + User.data.id,
			data: data,
			message: {
				succes: "user "+ User.data.name +" grant policy succes.",
				fatal: "user "+ User.data.name +" grant policy fatal.",
			}
		}).then(function() {
			window.history.back();
		})
	},
	removePolicy: function(id) {
		m.request({
			method: "DELETE",
			url: apiVersion + "/user/bind/policy/" + User.data.id + "/" + id,
			message: {
				succes: "remove policy succes",
				fatal: "remove policy fatal",
			}
		}).then(function() {
			User.bind.policy = User.bind.policy.filter(i => i.id!=id)
		})
	},
}

var Permission = {
	// model
	list: [],
	data: {id: 0, name: "",description: ""},
	table: new Tables("/permission/index", "/permission/count"),
	// view
	Index: {
		oninit: function(vnode) {
			Permission.table.size = 5
			Permission.table.head = ["id", "name", "description"]
			Permission.table.line = ["id", "name", "description"]
			Permission.table.setpack("name", function(i) {
				return m("a", {href: "#!/permission/info/" + i}, i)
			})
			Permission.table.redraw()
		},
		view: function(vnode) {
			return m("div", [
				m("div", [
					m("input#permission-search"),
					m("button", {onclick: Permission.searchPermission},"search key"),
					m("a", {href: "#!/permission/new"}, "new")
				]),
				m(Permission.table)
			])
		}
	},
	New: {
		view: function(vnode) {
			return m("div#permission-new", [
				m("div", [
					m("div", [m("label", "name"), m("input#permission-data-name")]),
					m("div", [m("label", "description"), m("input#permission-data-description")])
				]),
				m("button", {onclick: Permission.createPermissionById},"commit")
			])
		}
	},
	Edit: {
		oninit: function(vnode) {
			Permission.getPermissionById(vnode.attrs.name)
		},
		view: function(vnode) {
			return m("div#permission-edit", [
				m("div", [m("label", "id"), m("span", Permission.data["id"])]),
				m("div", [m("label", "name"), m("input#permission-data-name", {value: Permission.data["name"]})]),
				m("div", [m("label", "description"), m("input#permission-data-description", {value: Permission.data["description"]})]),
				m("button", {onclick: Permission.updatePermissionById},"commit")
			])
		}
	},
	Info: {
		oninit: function(vnode) {
			Permission.getPermissionById(vnode.attrs.name)
		},
		view: function(vnode) {
			return m("div#permission-info", [
				m("div", [m("label", "id"), m("span", Permission.data["id"])]),
				m("div", [m("label", "name"), m("span", Permission.data["name"])]),
				m("div", [m("label", "description"), m("span", Permission.data["description"])]),
				m("div", [
					m("a", {href: "#!/permission/edit/" + vnode.attrs.name}, "edit"),
					m("button", {onclick: Permission.deletePermissionById}, "delete")
				])
			])
		}
	},
	// controller
	getList: function() {
		m.request({
			method: "GET",
			url: apiVersion + "/permission/list"
		}).then(function(result){
			Permission.list = result
		})
	},
	getPermissionById(name) {
		m.request({
			method: "GET",
			url: apiVersion + "/permission/name/" + name
		}).then(function(result){
			Permission.data = result
		})
	},
	createPermissionById() {
		m.request({
			method: "PUT", 
			url: apiVersion + "/permission/new",
			data: {
				name: document.getElementById("permission-data-name").value,
				description: document.getElementById("permission-data-description").value,
			},
		})
	},
	updatePermissionById() {
		Permission.data.name = document.getElementById("permission-data-name").value
		Permission.data.description = document.getElementById("permission-data-description").value,
		m.request({
			method: "POST", 
			url: apiVersion + "/permission/id",
			data: {
				id: Permission.data.id,
				name: Permission.data.name,
				description: Permission.data.description
			},
		}).then(function() {
			window.history.back();
		})
	},
	deletePermissionById() {
		m.request({
			method: "DELETE",
			url: apiVersion + "/permission/id/" + Permission.data.id
		})
	},
	searchPermission() {
		var key = document.getElementById("permission-search").value
		if(key.length==0) {
			Permission.table.fetchData("/permission/index")
		}else {
			Permission.table.fetchData("/permission/search/" + key)
		}
	},
}

var Role = {
	data: {id: 0, name: "", description: ""},
	table: new Tables("/role/index"),
	bind: {
		user: [],
		permission: [],
	},
	grant: [],
	// view
	Index: {
		oninit: function(vnode) {
			Role.table.head = ["id", "name", "description"]
			Role.table.line = ["id", "name", "description"]
			Role.table.setpack("name", function(i) {
				return m("a", {href: "#!/role/info/" + i}, i)
			})
			Role.table.redraw()
		},
		view: function(vnode) {
			return m("div#role-index", [
				m("div", [
					m("input#role-search"),
					m("button", {onclick: Role.searchRole},"search key"),
					m("a", {href: "#!/role/new"}, "new")
				]),
				m(Role.table)
			])
		}
	},
	Info: {
		oninit: function(vnode) {
			Role.getRoleByName(vnode.attrs.name)
		},
		view: function(vnode) {
			return m("div#role-info", [
				m("div", [
					m("a", {href: "#!/role/edit/" + Role.data.name}, "edit"),
					m("button", {onclick: Role.deleteRole}, "delete"),
					m("a", {href: "#!/role/grant/user/" + Role.data.name}, "grant user"),
					m("a", {href: "#!/role/grant/permission/" + Role.data.name}, "grant permission"),
					m("button.button", {onclick: function() {window.history.back()} }, "cancel"),
				]),
				m("p", "基本信息"),
				m("p", Role.data.name),
				m("p", "权限描述"),
				m("p", Role.data.description),
				m("p", "权限列表"),
				m("div", Role.bind.permission.map(function(i){
					return m("tr", [
						m("td", m("a", {href: "#!/permission/info/" + i["name"]}, i["name"])),
						m("td", i.description),
						m("td", i.time),
						m("td", m("button", {onclick: function(){Role.removeRole(i["userid"])}}, "remove")),
					])
				})),
				m("p", "使用用户"),
				m("div", Role.bind.user.map(function(i){
					return m("tr", [
						m("td", m("a", {href: "#!/user/info/" + i["username"]}, i["username"])),
						m("td", i["time"]),
						m("td", m("button", {onclick: function(){Role.removeRole(i["userid"])}}, "remove")),
					])
				}))
			])
		}
	},
	New: {

		view: function(vnode) {
			return m("div#role-new", [
				m("p", "基本信息"),
				m("input#role-new-name"),
				m("p", "描述"),
				m("input#role-new-description"),
				m("div", [
					m("button.button", {onclick: Role.createRole}, "commit"),
					m("button.button", {onclick: function() {window.history.back()} }, "cancel"),
				])
			])
		}
	},
	Edit: {
		oninit: function(vnode) {
			Role.getRoleByName(vnode.attrs.name)
		},
		view: function(vnode) {
			return m("div#role-edit", [
				m("p", "基本信息"),
				m("input#role-edit-name", {value: Role.data.name}),
				m("p", "描述"),
				m("input#role-edit-description", {value: Role.data.description}),
				m("div", [
					m("button.button", {onclick: Role.updateRole}, "commit"),
					m("button.button", {onclick: function() {window.history.back()} }, "cancel"),
				])
			])
		}

	},
	GrantPermission: {
		view: function() {
			return m("div", "未实现")
		}
	},
	GrantUser: {
		view: function() {
			return m("div", "未实现")
		}
	},

	// controller
	getRoleByName: function(name) {
		m.request({
			method: "GET", 
			url: apiVersion + "/role/name/" + name,
		}).then(function(result){
			Role.data = result
			m.request({
				method: "GET", 
				url: apiVersion + "/role/user/id/" + result["id"],
			}).then(function(result){
				Role.bind.user = result || []
			})
			m.request({
				method: "GET", 
				url: apiVersion + "/role/permission/id/" + result["id"],
			}).then(function(result){
				Role.bind.permission = result || []
			})
		})

	},
	createRole: function() {
		m.request({
			method: "PUT",
			url: apiVersion + "/role/new",
			data: {
				name: document.getElementById("role-new-name").value,
				description: document.getElementById("role-new-description").value,
			},
			message: {
				succes: "create role " + document.getElementById("role-new-name").value + " succes!",
				fatal: "create role fatal."
			}
		})
	},

	updateRole: function() {
		Role.data.name = document.getElementById("role-edit-name").value
		Role.data.description = document.getElementById("role-edit-description").value
		m.request({
			method: "POST",
			url: apiVersion + "/role/id/" + Role.data.id,
			data: {
				name: Role.data.name,
				description: Role.data.description,
			},
			message: {
				succes: "update role " + Role.data.name + " succes!",
				fatal: "update role fatal."
			}
		}).then(function() {
			window.history.back();
		})
	},
	deletePolicy: function() {
		m.request({
			method: "DELETE",
			url: apiVersion + "/policy/id/" + Policy.data.id,
		}).then(function() {
			window.history.back();
		})
	},
}

var Policy = {
	data: {id: 0, name: "", policy: "{}", description: "",time: 0},
	table: new Tables("/policy/index", "/policy/count"),
	bind: [],
	grant: [],
	Index: {
		oninit: function(vnode) {
			Policy.table.head = ["id", "name", "description", "policy", "time"]
			Policy.table.line = ["id", "name", "description", "policy", "time"]
			Policy.table.setpack("name", function(i) {
				return m("a", {href: "#!/policy/info/" + i}, i)
			})
			Policy.table.setpack("time", function(i) {
				return new Date(i).Format("yyyy-MM-dd")
			})
			Policy.table.redraw()
		},
		view: function(vnode) {
			return m("div", [
				m("div", [
					m("input#policy-search"),
					m("button", {onclick: Policy.searchPolicy},"search key"),
					m("a", {href: "#!/policy/new"}, "new")
				]),
				m(Policy.table)
			])
		}
	},
	Info: {
		oninit: function(vnode) {
			Policy.getPolicyByName(vnode.attrs.name)
		},
		view: function(vnode) {
			return m("div#policy-info", [
				m("p", "基本信息"),
				m("p", Policy.data["name"]),
				m("p", Policy.data["description"]),
				m("p", "策略内容"),
				m("pre", JSON.stringify(JSON.parse(Policy.data["policy"]), null, 4)),
				m("div", [
					m("a", {href: "#!/policy/edit/" + Policy.data["name"]}, "edit"),
					m("button", {onclick: Policy.deletePolicy}, "delete"),
					m("a", {href: "#!/policy/grant/" + Policy.data["name"]}, "grant"),
					m("button.button", {onclick: function() {window.history.back()} }, "cancel"),
				]),
				m("p", "使用用户"),
				m("div", Policy.bind.map(function(i){
					return m("tr", [
						m("td", m("a", {href: "#!/user/info/" + i["username"]}, i["username"])),
						m("td", i["time"]),
						m("td", m("button", {onclick: function(){Policy.removePolicy(i["userid"])}}, "remove")),
					])
				}))
			])
		}
	},
	New: {
		view: function(vnode) {
			return m("div#policy-new", [
				m("p", "基本信息"),
				m("input#policy-new-name"),
				m("p", "描述"),
				m("input#policy-new-description"),
				m("p", "策略内容"),
				m("textarea#policy-new-policy"),
				m("div", [
					m("button.button", {onclick: Policy.createPolicy}, "commit"),
					m("button.button", {onclick: function() {window.history.back()} }, "cancel"),
				])
			])
		}
	},
	Edit: {
		oninit: function(vnode) {
			Policy.getPolicyByName(vnode.attrs.name)
		},
		view: function(vnode) {
			return m("div#policy-edit", [
				m("p", "基本信息"),
				m("input#policy-edit-name", {value: Policy.data["name"]}),
				m("p", "描述"),
				m("input#policy-edit-description", {value: Policy.data["description"]}),
				m("p", "策略内容"),
				m("textarea#policy-edit-policy", {value: JSON.stringify(JSON.parse(Policy.data["policy"]), null, 4)}),
				m("div", [

					m("button.button", {onclick: Policy.updatePolicy}, "commit"),
					m("button.button", {onclick: function() {window.history.back()} }, "cancel"),
				])
			])
		}
	},
	Grant: {
		oninit: function(vnode) {
			if(Policy.data.id==0) {

				Policy.getPolicyByName(vnode.attrs.name)
			}
			Policy.getListPolicy()
		},
		view: function(vnode) {
			return m("div#policy-grant", [
				m("div", Policy.data.name),
				m("div.box", [
					m("div", Policy.grant.map(function(i) {
						return !i.fouce ? m("tr", [
							m("td", i.name),
							m("td", i.description),
							m("td", m("button.button", {onclick:function(){i.fouce=true}}, "grant")),

						]) : ""
					})),
					m("div", Policy.grant.map(function(i) {
						return i.fouce ? m("tr", [
							m("td", i.name),
							m("td", i.description),
							m("td", m("button.button", {onclick:function(){i.fouce=false}}, "cancel")),
						]) : ""
					}))
				]),
				m("div", [
					m("button.button", {onclick: Policy.grantPolicy}, "commit"),
					m("button.button", {onclick: function() {window.history.back()} }, "cancel")
				]),
			])
		}
	},
	// controller
	getPolicyByName: function(name) {
		m.request({
			method: "GET", 
			url: apiVersion + "/policy/name/" + name,
		}).then(function(result){
			Policy.data = result
			m.request({
				method: "GET", 
				url: apiVersion + "/policy/user/id/" + result["id"],
			}).then(function(result){
				Policy.bind = result || []
			})
		})

	},
	getListPolicy: function() {
		m.request({
			method: "GET",
			url: apiVersion + "/policy/list",
		}).then(function(result) {
			Policy.grant = result || []
		})
	},
	createPolicy: function() {
		m.request({
			method: "PUT",
			url: apiVersion + "/policy/new",
			data: {
				name: document.getElementById("policy-new-name").value,
				description: document.getElementById("policy-new-description").value,
				policy: document.getElementById("policy-new-policy").value,
			},
			message: {
				succes: "create policy " + document.getElementById("policy-new-name").value + " succes!",
				fatal: "create policy fatal."

			}
		})
	},
	updatePolicy: function() {
		Policy.data.name = document.getElementById("policy-edit-name").value
		Policy.data.description = document.getElementById("policy-edit-description").value
		Policy.data.policy = document.getElementById("policy-edit-policy").value
		m.request({
			method: "POST",
			url: apiVersion + "/policy/id/" + Policy.data.id,
			data: {
				name: Policy.data.name,
				description: Policy.data.description,
				policy: Policy.data.policy,
			},
			message: {
				succes: "update policy " + document.getElementById("policy-edit-name").value + " succes!",
				fatal: "update policy fatal."
			}
		}).then(function() {
			window.history.back();
		})
	},
	deletePolicy: function() {
		m.request({
			method: "DELETE",
			url: apiVersion + "/policy/id/" + Policy.data.id,
		}).then(function() {
			window.history.back();
		})
	},
	removePolicy(id) {
		m.request({
			method: "DELETE",
			url: apiVersion + "/user/bind/policy/" + id +"/" + Policy.data.id
		})
	},
	grantPolicy: function() {
		var data = []
		for(var i of Policy.grant) {
			if(i.fouce) {
				data.push({id: i.id})
			}
		}
		if(data.length == 0) {
			return
		}

		m.request({
			method: "PUT",
			url: apiVersion + "/user/bind/policy/" + User.data.id,
			data: data,
			message: {
				succes: "user "+ User.data.name +" grant policy succes.",
				fatal: "user "+ User.data.name +" grant policy fatal.",
			}
		}).then(function() {
			window.history.back();
		})
	},
	searchPolicy() {
		var key = document.getElementById("policy-search").value
		if(key.length==0) {
			Policy.table.dataurl = "/policy/index"
		}else {
			Policy.table.dataurl = "/policy/search/" + key
		}
		Policy.table.redraw()
	},
}

var Nav = {
	menu: ["user", "permission", "role", "policy"],
	view: function() {
		return [
			m("nav#auth-nav.box",
				Nav.menu.map(function(i){
					return m("a", {href: "#!/" + i}, i)
				}
			)),
			m("div#auth-content.box")
		]
	}
}
window.onload = function() {
	m.mount(document.body, Home)
	m.mount(document.getElementById('container'), Nav)
	m.route(document.getElementById('auth-content'), "/user", {
		"/user": User.Index,
		"/user/info/:name": User.Info,
		"/user/edit/:name": User.Edit,
		"/user/grant/permission/:name": User.GrantPermission,
		"/user/grant/role/:name": User.GrantRole,
		"/user/grant/policy/:name": User.GrantPolicy,
		"/permission": Permission.Index,
		"/permission/new": Permission.New,
		"/permission/edit/:name": Permission.Edit,
		"/permission/info/:name": Permission.Info,
		"/role": Role.Index,
		"/role/new": Role.New,
		"/role/info/:name": Role.Info,
		"/role/edit/:name": Role.Edit,
		"/role/grant/permission/:name": Role.GrantPermission,
		"/role/grant/user/:name": Role.GrantUser,
		"/policy": Policy.Index,
		"/policy/new": Policy.New,
		"/policy/info/:name": Policy.Info,
		"/policy/edit/:name": Policy.Edit,
		"/policy/grant/:name": Policy.Grant,
	})
}

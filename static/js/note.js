"use strict";
var apiVersion = "/api/v1/note"

var Index = {
	data: [],
	oninit: function(vnode) {
		m.request({
			method: "GET", 
			url: apiVersion + "/index/list",
		}).then(function(result){
			Index.data = result || []
		})
	},
	view: function(vnode) {
		return m("div#note-container", [
			m("div#note-index", [
				m("ul", Index.data.map(function(i) {
					return m("li", [
						 m("a", {href: "#!/content/"+ i.path}, i.title),
						 m("p", i.content.slice(0,100)),
						 m("div", [
						 	m("span", i.createtime),
						 	m("span", i.ownerid)
						 ])
					])
				}))
			])
		])
		
	}
}

var User = {
	Index:{},
	Context: {},
}

var Context = {
	comment: [],
	Index: {
		oninit: function(vnode) {
			vnode.state.topics = []
			this.fetch(vnode)
		},
		onupdate: function(vnode) {
			if(vnode.attrs.path != vnode.state.path) {
				this.fetch(vnode)
			}
		},
		fetch: function(vnode) {
			vnode.state.path = vnode.attrs.path
			m.request({
				method: "GET", 
				url: apiVersion + "/content/" + vnode.attrs.path
			}).then(function(result){
				vnode.state.title = result.title
				vnode.state.time = result.edittime
				vnode.state.author = result.author
				vnode.state.topics = result.topics.split(",")
				if(result.format.trim() == "md") {
					vnode.state.content = marked(result.content)
				}else {
					vnode.state.content = result.content	
				}
				document.title = "eudore website note · " + vnode.state.title
				document.querySelector('meta[name="author"]').setAttribute("content", vnode.state.author)
				setTimeout(function(){
					Prism.highlightAll()
				}, 100)
			})

			m.request({
				method: "GET",
				url: apiVersion + "/comment/" + vnode.attrs.path
			}).then(function(result){
				Context.comment = result || []
				console.log(result)
			})
		},
		view: function(vnode) {
			return m('div#note-container', [
				// m('div.note-actions', [
				// 	m("li", m("a", {href: "#!/edit/" + vnode.attrs.path}, "edit")),
				// 	m("li", {onclick: ttt}, "share")
				// ]),
				m('div.note-user.row', [
					m("img.avatar", {src: "/api/v1/auth/user/icon/id/1"}),
					m("div.column", [
						m('span', "username"),
						m("div.row", [
							m('span.time',  new Date(vnode.state.time).Format("yyyy年M月d日")),
							m('span', 68)
						]),
						m("a.button", {href: "#!/edit/" + vnode.state.path},"edit")
					])
				]),
				m('div.note-header', [
					m('div.note-title', m('h1', vnode.state.title)),
					m('div.note-topics', vnode.state.topics.map(function(i) {
						return m('p', i)
					})),
					m('div.note-info', m("ul", [
						m("li.key", "作者："),
						m("li.val", vnode.state.author),
						m("li.key", "时间："),
						m("li", new Date(vnode.state.time).Format("yyyy年M月d日"))
					])),
				]),
				m('div.note-content', {innerHTML: vnode.state.content}),
				m('div.note-comment', Context.comment.map(function(i) {
					return m("div.row", [
						m("div", m("img.avatar", {src: "/api/v1/auth/user/icon/id/" + i.userid})),
						m("div.column", [
							m("span", i.name),
							m("span.time", i.createtime),
							m("p", i.content),
							m("div.row", [
								m("button", "consent"),
								m("button", "huifu")
							])
						]),
						
					])
				}))
			])
		}
	}
}

var Edit = {
	editor: null,
	oninit: function(vnode) {
		this.fetch(vnode)
	},
	onupdate: function(vnode) {
		if(vnode.attrs.path != vnode.state.path) {
			this.fetch(vnode)
		}
	},

	post: function(vnode) {
		console.log("edit")
		var content
		if(vnode.state.format == "md") {
			content = Edit.editor.value()
		}else if(vnode.state.format == "rich") {
			content = Edit.editor.txt.html()
		}
		m.request({
			method: "PUT",
			url: apiVersion + "/content/" + vnode.attrs.path,
			data: {
				title: vnode.state.title,
				topics: vnode.state.topics,
				format: vnode.state.format,
				content: content,
			},
		}).then(function(result){
			console.log(result)
		})
	},
	fetch: function(vnode) {
		vnode.state.path = vnode.attrs.path
		m.request({
			method: "GET", 
			url: apiVersion + "/content/" + vnode.attrs.path
		}).then(function(result){
			vnode.state.title = result.title
			vnode.state.topics = result.topics
			vnode.state.format = result.format.trim()
			vnode.state.content = result.content
		})
	},
	view: function(vnode) {
		return m('div#note-edit', [
			m("div#note-edit-header", [
				m("div.note-edit-title", [
					m("span", I18n("Title")),
					m("input.input", {type: "text", value: vnode.state.title})
				]),
				m("div", [
					m("span", I18n("Topics")),
					m("input.input", {type: "text", value: vnode.state.topics})
				]),
			]),
			m("div.note-content#editor",
				vnode.state.format == "rich" ?
				m("div#editor-rich", {
					oninit: function(){
						if(typeof window.Editor === "undefined"){
							loaddom("script", {
								type: "text/javascript",
								src: "/js/lib/wangEditor.js",
								// integrity: "sha512-2eA+lwf7NEa43PE5A0E61L85tdSOWyaVCS0LnWMO/8SxmfTwEZmUHmWFT7lbbvBUu4Ym3oBLPiPxtT61dKbZ8g=="
							})
						}
						Edit.timer = setInterval(function(){
							if(typeof window.wangEditor === "undefined"){
								return
							}
							clearInterval(Edit.timer)
							Edit.editor = new window.wangEditor('#editor-rich')
							Edit.editor.customConfig.uploadImgShowBase64 = true
							// editor 英文化
							if(Base.lang=='en-US') {
								Edit.editor.customConfig.lang = {
									'设置标题': 'Title',
									'正文': 'p',
									'文字颜色': 'Text Color',
									'背景色': 'Background Color',
									'链接文字': 'Text Link',
									'设置列表': 'Setting List',
									'有序列表': 'Order',
									'无序列表': 'Unorder',
									'对齐方式': 'Alignment',
									'靠左': 'Left',
									'居中': 'Center',
									'靠右': 'Right',

									'上传图片': 'Upload Image',
									'网络图片': 'Web Image',
									'图片链接': 'Image Link',

									'插入视频': 'Insert Video',
									'格式如': 'Format such as',
									'插入代码': 'Insert Code',
									'显示行号': 'Display line number',
									'高亮语法': 'Highlight Grammar',

									'创建': 'Init',
									'行': 'Row',
									'列的表格': "Column Table",
									'表格': 'Table',
									'上传': 'Upload',
									'插入': 'Insert',
								}
							}
							Edit.editor.create()
							Edit.editor.txt.html(vnode.state.content)
						}, 30)
					}
				}) : "",
				vnode.state.format == "md" ? 
				m("div#editor-md", m("textarea", {
					oninit: function(){
						if(typeof window.SimpleMDE === "undefined"){
							loaddom("script", {
								type: "text/javascript",
								src: "/js/lib/simplemde.min.js",
								// integrity: "sha512-ksSfTk6JIdsze75yZ8c+yDVLu09SNefa9IicxEE+HZvWo9kLPY1vrRlmucEMHQReWmEdKqusQWaDMpkTb3M2ug=="
							})
							loaddom("link", {
								rel: "stylesheet",
								href: "/css/lib/simplemde/simplemde.min.css",
								// integrity: "sha512-No/9S1HC+I/G1jzh0zUXy0HLTHxgyDolWOHui0KAjpXExsIbtJDbt0f6P6vy4x60GoNo95KfkovR0GaK6/b8ig=="
							})
						}
						Edit.timer = setInterval(function(){
							if(typeof window.SimpleMDE === "undefined"){
								return
							}
							clearInterval(Edit.timer)

							Edit.editor = new window.SimpleMDE({
								element: document.getElementById('editor-md').querySelector('textarea'),
								showIcons: ["code", "table", "undo", "redo"],
							})
							Edit.editor.value(vnode.state.content)
							// i18n SimpleMDE title
							for(var i of document.querySelectorAll('.editor-toolbar a')) {
								var pos = i.title.indexOf(' (')
								if(pos == -1) {
									pos = i.title.length
								}
								i.title = I18n(i.title.slice(0, pos)) + i.title.slice(pos)
							}
						}, 30)
					}
				}
			)) : "",
			),
			m("div#note-edit-commit", [
				m("button", {onclick: function(){
					Edit.post(vnode)
				}}, "commit")
			])
		])
	}
}


window.onload = function() {
	m.mount(document.body, Home)
	m.route(document.getElementById('container'), "/index", {
		"/index": Index,
		"/index/:path...": User.Index,
		"/content/:path...": Context.Index,
		"/edit/:path...": Edit,
	})
}


function ttt() {
	for(var i=0;i<10;i++) {

		m.request({
			method: "POST",
			url: apiVersion + "/note/content" + uri,
			data: {a:1,b:2}
		}).then(function(result){
			console.log(result)
		})
	}
}

// Message.add("执行命令：server成功。")
// Message.add('添加更新文档“Home - Jass”成功！')
// Message.add('添加更新文档“Home - Jass”成功！', "warring")
// Message.add('添加更新文档“Home - Jass”成功！\n11111111111111', "fatal")
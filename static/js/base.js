"use strict";
// Array Remove - By John Resig (MIT Licensed)
Array.prototype.remove = function(from, to) {
	var rest = this.slice((to || from) + 1 || this.length);
	this.length = from < 0 ? this.length + from : from;
	return this.push.apply(this, rest);
};
// (new Date()).Format("yyyy-MM-dd hh:mm:ss.S") ==> 2006-07-02 08:09:04.423 
// (new Date()).Format("yyyy-M-d h:m:s.S")      ==> 2006-7-2 8:9:4.18 
var Month = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
Date.prototype.Format = function (fmt) {  
	var o = {
		"A+": Month[this.getMonth()],
		"M+": this.getMonth() + 1, //月份 
		"d+": this.getDate(), //日 
		"h+": this.getHours(), //小时 
		"m+": this.getMinutes(), //分 
		"s+": this.getSeconds(), //秒 
		"q+": Math.floor((this.getMonth() + 3) / 3), //季度 
		"S": this.getMilliseconds() //毫秒 
	};
	if (/(y+)/.test(fmt)) fmt = fmt.replace(RegExp.$1, (this.getFullYear() + "").substr(4 - RegExp.$1.length));
	for (var k in o)
	if (new RegExp("(" + k + ")").test(fmt)) fmt = fmt.replace(RegExp.$1, (RegExp.$1.length == 1) ? (o[k]) : (("00" + o[k]).substr(("" + o[k]).length)));
	return fmt;
};


window.onerror = function(errorMessage, scriptURI, lineNo, columnNo, error) {
    console.log('message: ' + errorMessage); // 异常信息
    console.log('file: ' + scriptURI); // 异常文件路径
    console.log('line: ' + lineNo); // 异常行号
    console.log('column: ' + columnNo); // 异常列号
    console.log('error: ' + error); // 异常堆栈信息
};

function loaddom(tag, args) {
	var dom = document.createElement(tag)
	for(var key in args) {
		dom.setAttribute(key, args[key])
	}
	document.body.appendChild(dom);
}

function I18n(key) {
	return langs[Base.lang][key] || key
}

var langs = {
	"en-US": {},
	"zh-CN": {
		"history": "历史",
		"help": "帮助",
		"about": "关于",
		"setting": "设置",
		// auth
		'user': '用户',
		'permission': '权限',
		'role': '角色',
		'policy': '策略',
		// auth login
		'username': '用户名',
		'password': '密码',
		'captcha': '验证码',
		// file
		'Overview': '概览',
		'Files': '文件管理',
		'Basic Settings': '基础设置',
		'Log Overview': '日志查询',
		'Create Folder': '新建目录',
		'Upload File': '上传文件',
		'Upload Folder': '上传目录',
		// note
		// note-edit-md
		'Bold': '加粗',
		'Italic': '斜体',
		'Heading': '标题',
		'Code': '代码',
		'Quote': '引用',
		'Generic List': '通用列表',
		'Numbered List': '编号列表',
		'Create Link': '创建链接',
		'Insert Image': '插入图片',
		'Insert Table': '插入表格',
		'Toggle Preview': '预览',
		'Toggle Side by Side': '并排显示',
		'Toggle Fullscreen': '全屏',
		'Markdown Guide': 'markdown指南',
		'Undo': '撤销',
		'Redo': '反撤销',
		// status-build
		'Build Info': '编译信息',
		"Server Language": "服务端语言",
		"Server Web Framework": "服务端web框架",
		"Server Database": "服务端数据库",
		"Frotend Web Framework": "前端web框架",
		"Frotend Highlighting": "前端高亮",
		"Frotend Markdown": "前端Markdown",
		// status system
		'System Info': '系统信息',
		'Uptime': "服务运行时间",
		'NumGoroutine':	'当前 Goroutines 数量',
		'MemAllocated': '当前内存使用量',
		'MemTotal': '所有被分配的内存',
		'MemSys': '内存占用量',
		'Lookups': '指针查找次数',
		'MemMallocs': '内存分配次数',
		'MemFrees': '内存释放次数',
		'HeapAlloc': '当前 Heap 内存使用量',
		'HeapSys': 'Heap 内存占用量',
		'HeapIdle': 'Heap 内存空闲量',
		'HeapInuse': '正在使用的 Heap 内存',
		'HeapReleased': '被释放的 Heap 内存',
		'HeapObjects': 'Heap 对象数量',
		'StackInuse': '启动 Stack 使用量',
		'StackSys': '被分配的 Stack 内存',
		'MSpanInuse': 'MSpan 结构内存使用量',
		'MSpanSys': '被分配的 MSpan 结构内存',
		'MCacheInuse': 'MCache 结构内存使用量',
		'MCacheSys': '被分配的 MCache 结构内存',
		'BuckHashSys': '被分配的剖析哈希表内存',
		'GCSys': '被分配的 GC 元数据内存',
		'OtherSys': '其它被分配的系统内存',
		'NextGC': '下次 GC 内存回收量',
		'LastGC': '距离上次 GC 时间',
		'PauseTotalNs': 'GC 暂停时间总量',
		'PauseNs': '上次 GC 暂停时间',
		'NumGC': 'GC 执行次数',
		'Config Info': '配置信息',
	}
}


var _self = (typeof window !== 'undefined')
	? window   // if in browser
	: (
		(typeof WorkerGlobalScope !== 'undefined' && self instanceof WorkerGlobalScope)
		? self // if in worker
		: {}   // if in node js
	);

_self.Base = JSON.parse('{"userid":0,"name":"","lang":"en-US","level":0,"menu":0,"bearer":""}')


if(window.localStorage){
	var userinfo = JSON.parse(localStorage.getItem("user") || '{}')
	if(new Date().getTime() < (userinfo.expires || 0) * 1000) {
		_self.Base = userinfo
		// _self.Base.bearer = localStorage.getItem("user")
		console.log("load ")
	}else {
		localStorage.setItem("user", '{}')
	}
	if(typeof Base.lang === 'undefined') {
		var lang = ""
		if(document.querySelector('meta[name="language"]')){
			lang = document.querySelector('meta[name="language"]').getAttribute('content')
		}
		for(var i of lang.split(',')) {
			var pos = i.indexOf(';')
			if(pos != -1) {
				i = i.slice(0, pos)
			}
			if(i in langs) {
				Base.lang = i
				break
			}
		}
	}
	if(typeof Base.lang === 'undefined') {
		Base.lang = "en-US"
	}
	// _self.Base = JSON.parse(localStorage.getItem("user"))
	// _self.Base = JSON.parse('{"userid":1,"name":"root","lang":"en-US","level":0,"menu":255,"bearer":"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcmVzIjoxNjY1NDAzMjAwLCJuYW1lIjoicm9vdCIsInVzZXJpZCI6IjEifQ.N7dACszdw1iScsTGYCv8EPrbFCqxcow0gvC_uKbidC0"}')
	// _self.Base = JSON.parse('{"Name":"user01","Level":0,"menu":255,"Bearer":"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NjU0MDMyMDAsIm5hbWUiOiJ1c2VyMDEiLCJ1aWQiOiIzIn0.Ec7ECqsUxIZJvts46C0zVLiMzErqKzL7TUdMLQ0Qpk4"}')
}

//
/*if(window.parent!=null) {
	if(document.querySelector('meta[name="parent-id"]') == null) {
		var pid = document.createElement('meta')
		pid.name= "parent-id"
		document.querySelector("head").appendChild(pid)
	}
	document.querySelector('meta[name="parent-id"]').setAttribute("content", window.parent.document.querySelector('meta[name="request-id"]').getAttribute('content'))
}*/

// 

var requestid = ""
if(document.querySelector('meta[name="request-id"]') != null) {
	requestid = document.querySelector('meta[name="request-id"]').getAttribute('content')
}
var parser = document.createElement('a');
if(typeof m.request !== 'undefined') {
	var oldrequest = m.request
	m.request = function(args) {
		var extract = args.extract
		var message = args.message || {}
		delete args.message
		args.extract = function(xhr) {
			// console.log(xhr, xhr.status)
			if(xhr.status >= 400){
				parser.href = xhr.responseURL
				Message.add('request ' + parser.pathname + " status is " + xhr.status, "fatal")
				if(message.fatal != "" && typeof  message.fatal !== "undefined") {
					Message.add(message.fatal, "fatal")
				}
				return new Promise(function() {})
			}else {
				if(message.succes != "" && typeof  message.succes !== "undefined") {
					Message.add(message.succes, "succes")
				}
			}
			return extract ? extract(xhr) : JSON.parse(xhr.responseText || '{}')
		}

		// add header
		if(!("headers" in args)) {
			args["headers"] = {}
		}
		if(Base.lang != "") {
			args["headers"]["Accept-Language"] = Base.lang
		}
		if(Base.bearer != "") {
			args["headers"]["Authorization"] = Base.bearer
		}
		if(requestid!="") {
			args["headers"]["X-Parent-Id"] = requestid
		}
		
		return oldrequest(args)
	}
}

function allxhr(xhr) {
	// Get the raw header string
	var headers = splitheaders(xhr.getAllResponseHeaders());

	return {xhr: xhr, status: xhr.status, body: xhr.response, header: headers}
}

function splitheaders(headers) {
	// Convert the header string into an array
	// of individual headers
	var arr = headers.trim().split(/[\r\n]+/);

	// Create a map of header names to values
	var headerMap = {};
	arr.forEach(function (line) {
		var parts = line.split(': ');
		var header = parts.shift();
		var value = parts.join(': ');
		headerMap[header] = value;
	});
	return headerMap
}
function parseJwt(token) {
	var base64Url = token.split('.')[1];
	var base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
	return JSON.parse(window.atob(base64));
};


const Header = {
	nav: ["auth", "note", "file", "status"],
	view: function() {
		return m('header.header', [
			m("div.left", [
				m("div.header-icon", m("a", {href: "/"}, m("img", {src: "/favicon.ico", alt: "wejass.com", height: 20, width: 20}))),
				m("div.header-nav", m("ul", Header.nav.map(function(k){
					return m("li", m("a", {href: "/" + k + "/"}, k))
				})))
			]),
			m("div.right", [
				m("div.header-search", 
					m("input", {type: "text", placeholder: "Search"})
				),
				m("div.header-user", 
					Base.name == "" ? m("a", {href: "/auth/login/wejass?location=" + window.location.href}, "Log in") : [
						m("img.avatar", {src: "/api/v1/auth/user/icon/name/" + Base.name, onclick: Base.showPage}),
						m("ul.user-nav", [
							m("li", Base.name),
							m("li.divider"),
							// m("li", "My Account"),
							m("li", m("a", {href: "/auth/user/setting"}, "Setting")),
							m("li", m("a", {href: "/auth/user/logout"}, "Log Out"))
						])
					]
				)
			])
		])
	}
}



const Content = {
	view: function() {
		return m('main#container')
	}
}

const Footer = {
	footer: [
		{name: "history", url: "/eudore/history"},
		{name: "help", url: "/eudore/history"},
		{name: "about", url: "/eudore/history"},
		{name: "setting", url: "/eudore/setting"},
	],
	view: function() {
		return m("footer.footer", [
			m("div.footer-link.column", m("ul", Footer.footer.map(function(u){ 
				return m("li", m("a", {href: u.url}, I18n(u.name)))
			}))),
			m("div", "Version: 0.0.0")
		])
	}
}

const Message = {
	list: [],
	time: null,
	add: function(msg, state) {
		Message.list.push({
			msg: msg,
			state: state ? "." + state : ".succes",
			time: 5000 + (+new Date())
		})
		m.redraw()
		if(Message.time == null) {
			Message.time = setInterval(Message.ontime, 100)
		}
	},
	ontime: function() {
		var list = Message.list
		for(var i in list) {
			if(list[i]["time"] < +new Date()) {
				list.shift()
				m.redraw()
			}else {
				break
			}
		}
		if(list.length==0) {
			clearInterval(Message.time)
			Message.time = null
		}
	},
	view: function() {
		return m("div#message", Message.list.map(function(i){
			return m("div" + i.state,  m("span", i.msg))
		}))
	}
}

const Side = {
	view: function() {
		return m("div#side-tool")
	}
}
const Home = {
	view: function(v) {
		return [
			m(Header),
			m(Message),
			m(Content),
			m(Side),
			m(Footer),
		]
	}
}


_self.Message = Message

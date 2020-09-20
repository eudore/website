"use strict";
// a.remove("tag") 删除指定值
Array.prototype.remove = function() {
    var what, a = arguments, L = a.length, ax;
    while (L && this.length) {
        what = a[--L];
        while ((ax = this.indexOf(what)) !== -1) {
            this.splice(ax, 1);
        }
    }
    return this;
};
// 不重复追加值
Array.prototype.pushNoRepeat = function() {
	for (var i = 0; i < arguments.length; i++) {
		var ele = arguments[i];
		if (this.indexOf(ele) == -1) {
			this.push(ele);
		}
	}
};
// (new Date()).Format("yyyy-MM-dd hh:mm:ss.S") ==> 2006-07-02 08:09:04.423 
// (new Date()).Format("yyyy-M-d h:m:s.S")      ==> 2006-7-2 8:9:4.18 
var Month = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
Date.prototype.Format = function(fmt) {
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



var Base = {
	"userid": 0,
	"name": "",
	"lang": "",
	"bearer": ""
}
if (window.localStorage) {
	var userinfo = JSON.parse(localStorage.getItem("user") || '{}')
	if (new Date().getTime() < (userinfo.expires || 0) * 1000) {
		Base = userinfo
	} else {
		localStorage.setItem("user", '{}')
	}
}
if (typeof Base.lang === 'undefined' || Base.lang == "") {
	Base.lang = "en-US"
}


var apiVersion = "/api/v1"
var requestid = ""
if (document.querySelector('meta[name="request-id"]') != null) {
	requestid = document.querySelector('meta[name="request-id"]').getAttribute('content')
}
axios.interceptors.request.use(
	function(config) {
		if (config.url.indexOf('/') != 0 && config.url.indexOf('http://') != 0 && config.url.indexOf('https://') != 0 ) {
			config.url = apiVersion + '/' + config.url
		}
		if (!("timeout" in config)) {
			config["timeout"] = 4000
		}
		// add header
		if (!("headers" in config)) {
			config["headers"] = {}
		}
		if (Base.lang != "") {
			config["headers"]["Accept-Language"] = Base.lang
		}
		if (Base.bearer != "") {
			config["headers"]["Authorization"] = Base.bearer
		}
		if (requestid != "") {
			config["headers"]["X-Parent-Id"] = requestid
		}
		return config
	}
)

function getRoutes(prefix, data) {
	var a = []
	if (data["template"]) {
		a.push({
			path: prefix || data["path"],
			component: data,
		})
	}
	for (var i in data) {
		if (/^[A-Z][a-z]+$/.test(i)) {
			a = a.concat(getRoutes(prefix + (data[i]["path"] || "/" + i.toLowerCase()), data[i]))
		}
	}
	return a
}


Vue.component("App", {
	template: `
	<div id="app">
		<Header></Header>
		<Container></Container>
		<Side/>
		<Footer></Footer>
	</div>
	`
})
Vue.component("Header", {
	data: function() {
		return {
			// navs: ["auth", "file", "note", "chat", "term", "task", "status"],
			navs: ["auth", "note", "term", "status"],
		}
	},
	methods: {
		logout: function() {Base.userid = 0; Base.name = ""; Base.bearer = ""; localStorage.setItem("user", '{}'); }
	},
	computed: {
		getLoginLocation: function() {
			console.log('/auth/login?location=' + window.location.href.slice(window.location.origin.length))
			return '/auth/login?location=' + window.location.href.slice(window.location.origin.length)
		}
	},
	template: `
	<header class="header">
		<div id="header-bar">
			<div id="header-icon">
				<a href="/"><img src="/favicon.ico" alt="/"></a>
			</div>
			<div id="header-nav">
				<ul>
					<li v-for="nav in navs">
						<a :href="'/'+nav+'/'">{{nav}}</a>
					</li>
				</ul>
			</div>
			<div id="header-search">
				<input type="text" placeholder="Search" id="header-search-input">
				<div id="header-search-list"></div>
			</div>
			<div id="header-user">
				<img v-if="$root.Base.name" :src="'/api/v1/auth/icon/'+$root.Base.userid" class="avatar">
				<a v-else :href="getLoginLocation">Log in</a>
					<ul v-if="$root.Base.name" id="header-user-nav">
						<li><a>{{ $root.Base.name }}</a></li>
						<li class="divider"></li>
						<li><a href="/auth/setting">Setting</a></li>
						<li @click="logout">Log Out</li>
					</ul>
				</div>
			</div>
		</div>
	</header>
	`
})
Vue.component("Footer", {
	template: `
	<footer class="footer">
		<div id="footer-bar">
			<div>
				<ul id="footer-links">
					<li><a href="/eudore/history">历史</a></li>
					<li><a href="/eudore/history">帮助</a></li>
					<li><a href="/eudore/history">关于</a></li>
					<li><a href="/eudore/setting">设置</a></li>
				</ul>
			</div>
			<div> BuildTime: 2020-04-04_08:16:58 CommitID: </div>
		</div>
	</footer>
	`
})
Vue.component("Container", {
	template: `
	<main id="container">
		<nav id="nav-menu" class="box">
			<router-link v-for="nav in $root.navs" :to="'/'+nav.toLowerCase()+'/index'" :key="nav">{{nav}}</router-link>
		</nav>
		<router-view></router-view>
	</main>
	`
})
Vue.component("Side", {
	methods:{
		totop: function() {
			let top = document.documentElement.scrollTop || document.body.scrollTop;
			let sub = top/10
			if(sub<50){
				sub=50
			}
			const timeTop = setInterval(() => {
				document.body.scrollTop = document.documentElement.scrollTop = top -= sub;
				if (top <= 0) {
					clearInterval(timeTop);
				}
			}, 10);
		}
	},
	template:`<div id="side-tool">
		<div @click="totop">
<svg aria-hidden="true" focusable="false" data-prefix="fas" data-icon="chevron-circle-up" class="svg-inline--fa fa-chevron-circle-up fa-w-16" role="img" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
    <path fill="currentColor" d="M8 256C8 119 119 8 256 8s248 111 248 248-111 248-248 248S8 393 8 256zm231-113.9L103.5 277.6c-9.4 9.4-9.4 24.6 0 33.9l17 17c9.4 9.4 24.6 9.4 33.9 0L256 226.9l101.6 101.6c9.4 9.4 24.6 9.4 33.9 0l17-17c9.4-9.4 9.4-24.6 0-33.9L273 142.1c-9.4-9.4-24.6-9.4-34 0z">
    </path>
</svg>
		</div>
	</div>`,
})

Vue.component("TableController-Data", {
	data: function() {
		return {
			page: 0, size: 20, count: 0, datas: [],
			searchkey: "",_url:"", body: "", 
		}
	},
	props: ['name','url', 'value','tableFields', 'extraFields'],
	mounted: function() {
		this.load(0)
	},
	methods: {
		load: function(n) {
			if (n < 0 || n > this.count / this.size) {return }
			if (this.value) {
				this.count = this.value.length
				this.datas = this.value.slice(n * this.size, (n + 1) * this.size)
				this.page = n
				return
			}
			if(this.name){this._url = this.dataurl() }
			if(this.url){this._url=this.url }
			if (this._url) {
				var req = {
					method: this.body ? "post" : "get",
					url: this._url, data: this.body,
					params: {page: n, size: this.size, },
				}
				axios.request(req).then((response) => {
					if ("count" in response.data) {
						this.count = response.data.count;
						this.datas = response.data.data
					} else {
						this.count = 1;
						this.datas = [response.data]
					}
					this.page = n
				})
			}
		},
		dataurl: function() {
			this.body = null
			if (this.searchkey == "") {
				return this.name + '/list'
			}
			var conds = this.parseSearchExpression(this.searchkey)
			console.log(conds)
			if (conds.length == 0) {
				return this.name + '/search/default/' + this.searchkey
			} else if (conds.length == 1) {
				if(conds[0].field=="id"||conds[0].field=="name"){
					return this.name + '/'+ conds[0].field + '/' + conds[0].value
				}
				return this.name + '/search/' + conds[0].field + '/' + conds[0].value
			} else {
				this.body = conds
				return this.name + '/search'
			}
		},
		parseSearchExpression: function(key) {
			var keys = key.split(/(\S+\'.*\'|\S+)/).filter(key => /\S+/.test(key))
			var data = []
			for (var i of keys) {
				var exp = i.split(/(\w*)(=|>|>|<>|!=|>=|<=|~|!~)(.*)/).filter(key => /\S+/.test(key))
				if (exp.length != 3) {
					continue
				}
				if (exp[1] == "~") {
					exp[1] = "LIKE"
				} else if (exp[1] == "!~") {
					exp[1] = "NOT LIKE"
				}
				data.push({
					field: exp[0],
					expression: exp[1],
					value: exp[2],
				})
			}
			return data
		},
	},
	computed: {
		pagelist: function() {
			var num = this.count / this.size
			var pages = []
			for (var i = this.page - 3; i < num; i++) {
				if (i >= 0) {
					pages.push(i)
				}
			}
			return pages
		},
	},
	template: `
	<div class="table-data">
		<input v-if="name" v-model="searchkey" @input="load(0)"></input>
		<table :key="searchkey.length">
			<tr>
				<th v-for="i in tableFields">{{i}}</th>
				<template v-for="i in extraFields">
					<th v-if="i">{{ i }}</th>
				</template>
			</tr>
			<tr v-for="data in datas">
				<td v-for="i in tableFields"">
					<template v-if="$scopedSlots[i]"><slot :name="i" :value="data[i]" :data="data"/></template>
					<template v-else>{{data[i]}}</template>
				</td>
				<template v-for="i in extraFields">
					<td v-if="$scopedSlots[i]"><slot :name="i" :value="data[i]" :data="data" /></td>
				</template>
			</tr>
		</table>
		<div>
			<ul>
			<li @click="load(page-1)">pre</li>
			<li v-for="i in pagelist"><a class="button" @click="load(i)"  >{{i+1}}</a></li>
			<li @click="load(page+1)">next</li>
			</ul>
		</div>
	</div>
	`
})

Vue.component("TableController-Index", {
	props: ['name', 'extraFields'],
	template: `
	<div>
		<router-link :to="'/'+name+'/new'" class="button">new</router-link>
		<TableController-Data v-cloak :name="name" :tableFields="$root.fields[name]" :extraFields="extraFields">
			<template v-slot:id="props">
				<router-link :to="'/'+name+'/info/'+props.value">{{ props.value }}</router-link>
			</template>
			<template v-slot:name="props">
				<router-link :to="'/'+name+'/info/'+props.value">{{ props.value }}</router-link>
			</template>
			<template v-slot:time="props">
				{{ (new Date(props.value )).Format("yyyy-MM-dd hh:mm") }}
			</template>
			<template v-slot:logintime="props">
				{{ (new Date(props.value )).Format("yyyy-MM-dd hh:mm") }}
			</template>
			<template v-slot:sigintime="props">	{{ (new Date(props.value)).Format("yyyy-MM-dd hh:mm") }} </template>
			<template v-for="i in Object.keys($scopedSlots)" v-slot:[i]="props"> 
				<slot :name="i" v-bind=props></slot> 
			</template>
		</TableController-Data>
	</div>
	`,
})

Vue.component("TableController-New", {
	data: function() {
		return {
			field: ['id', 'name'],
			value: {},
		}
	},
	props: ['name'],
	mounted: function() {
		this.init()
	},
	methods: {
		init: function() {
			axios.get(apiVersion + "/" + this.name + '/fields').then((response) => {
				this.field = response.data.slice(1)
			})
		},
		commit: function() {
			console.log("new", this.field, this.value)
			axios.post(apiVersion + "/" + this.name + '/new', this.value).then((response) => {
				console.log(response)
			})
		},
	},
	template: `
	<div>
		<table>
			<tbody>
				<tr v-for="i in field">
					<td>{{i}}</td>
					<td><input v-model=value[i]></input></td>
				</tr>
			</tbody>
		</table>
		<button @click="commit">commit</button>
		<button @click="$root.$router.go(-1)">cannel</button>
	</div>
	`
})

Vue.component("TableController-Info", {
	data: function() {
		return {
			data: {},
			field: ['id', 'name'],
			value: {},
		}
	},
	props: ['name'],
	mounted: function() {
		this.init()
	},
	methods: {
		init: function() {
			var key = this.$route.params.key
			if (/^\d+$/.test(key)) {
				axios.get(apiVersion + "/" + this.name + '/id/' + key).then((response) => {
					this.data = response.data
				})
			} else {
				axios.get(apiVersion + "/" + this.name + '/name/' + key).then((response) => {
					this.data = response.data
				})
			}

		},
	},
	template: `
	<div>
		<table>
			<tbody>
				<tr v-for="i in Object.keys(data)">
					<td>{{i}}</td>
					<td>{{ data[i] }}</td>
				</tr>
			</tbody>
		</table>
		<slot name="info-bar"/>
	</div>
	`
})

Vue.component("TableController-Bind", {
	data: function() {
		return {
			listdata: [],
			havedata: [],
			alldata: [],
			selectdata: [],
		}
	},
	props: ['source','target', 'listurl', 'haveurl', 'commiturl', 'grantFields', 'tableFields'],
	mounted: function() {
		this.init() 
	},
	methods: {
		init: function() {
			axios.get(this.listurl).then((response)=>{this.listdata=response.data.data;this.filterhave();})
			if (this.haveurl) {
				axios.get(this.$root.getRouteUrl(this.haveurl)).then((response)=>{this.havedata=response.data.data||[];this.filterhave();})
			}
		},
		filterhave: function() {this.alldata=this.listdata.filter((v)=>{for(var i of this.havedata){if(v.id==i[this.target+'id']){return false}}return true})},
		select: function(val, effect) {val.effect=effect;this.selectdata.pushNoRepeat(val)},
		unselect: function(val) {this.selectdata=this.selectdata.filter((v)=>{return v.id != val.id})},
		commit: function() {
			var datas = []
			for (var i of this.selectdata) {
				var data = {};data[this.source+"id"]=this.$route.params.key;data[this.target+"id"]=i.id
				if (i.effect !== undefined) {data.effect = i.effect}
				datas.push(data)
			}
			axios.put(this.commiturl, datas)
		},
	},
	template: `
	<div class="bind">
		<div> {{ source }} > Bind > {{ target }}</div>
		<div class="row">
			<TableController-Data :key="'a'+this.alldata.length" :name="target" :tableFields="grantFields" >
				<template v-slot:allow="props">
					<button @click="select(props.data,true)"> allow</button>
				</template>
				<template v-slot:deny="props">
					<button @click="select(props.data,false)"> deny</button>
				</template>
				<template v-slot:grant="props">
					<button @click="select(props.data)"> grant</button>
				</template>
			</TableController-Data>
			<TableController-Data :key="'s'+this.selectdata.length" :value="this.selectdata" :tableFields="['id','name','cannel']">
				<template v-slot:cannel="props">
					<button @click="unselect(props.data)">cannel</button>
				</template>
			</TableController-Data>
		</div>
		<div>
			<button @click="$root.$router.go(-1)">cannel</button>
			<button @click="commit">commit</button>
		</div>
	</div>
	`,
})


function I18n(key) {
	return langs[Base.lang][key] || key
}

var langs = {
	"en-US": {},
	"zh-CN": {
		"eudore website is a multifunction platform.": "eudore网站是一个多功能平台。",
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
		'NumGoroutine': '当前 Goroutines 数量',
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

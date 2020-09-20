//


apiVersion = "/api/v1/status"

Vue.component("Container", {
	data: function() {
		return {
			build: [],
			system: {},
			systeminfo: ['Uptime', 'NumGoroutine', 'MemAllocated', 'MemTotal', 'MemSys', 'Lookups', 'MemMallocs', 'MemFrees', 'HeapAlloc', 'HeapSys', 'HeapIdle', 'HeapInuse', 'HeapReleased', 'HeapObjects', 'StackInuse', 'StackSys', 'MSpanInuse', 'MSpanSys', 'MCacheInuse', 'MCacheSys', 'BuckHashSys', 'GCSys', 'OtherSys', 'NextGC', 'LastGC', 'PauseTotalNs', 'PauseNs', 'NumGC'],
			config: [],
		}
	},
	mounted: function() {
		this.init()
	},
	methods: {
		init: function() {
			axios.get(apiVersion + "/build").then((response) => {
				this.build = response.data
			})
			axios.get(apiVersion + "/system").then((response) => {
				this.system = response.data
			})
			axios.get(apiVersion + "/config").then((response) => {
				this.config = response.data
			})
		},
	},
	template: `
	<main id="container">
		<div id="status-build">
			<fieldset>
				<legend>{{$t('Build Info')}}</legend>
				<tr v-for="i in build">
					<td>{{$t(i.name)}}</td><td><a :href="i.link">{{i.version}}</a></td>
				</tr>
			</fieldset>
		</div>
		<div id="status-system">
			<fieldset>
				<legend>System Info</legend>
				<tr v-for="i in systeminfo">
					<td>{{$t(i)}}</td><td>{{$t(system[i])}}</td>
				</tr>
			</fieldset>
		</div>
		<div id="status-config">
			<fieldset>
				<legend>Config Info</legend>
				<tr v-for="i in config">
					<td>{{i.name}}</td><td>{{i.value}}</td>
				</tr>
			</fieldset>
		</div>

	</main>
	`,
})
document.body.className="status"


// 准备翻译的语言环境信息
const messages = {
  en: {
  },
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

// Base.lang='zh-CN'
// 通过选项创建 VueI18n 实例
const i18n = new VueI18n({
  locale: Base.lang, // 设置地区
  messages, // 设置地区信息
})

var vm = new Vue({
	el: '#app',
	template: `<App></App>`,
	data: function() {
		return {
			Base: Base,
		}
	},
	i18n,
})
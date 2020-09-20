
//
apiVersion = "/api/v1/note"

Vue.component("Container", {
	template: `
	<main id="container">
		<router-view></router-view>
	</main>
	`
})

var treekey = 0
function Tree(){
	this.data ={}
	this.childs={}
	this.key = 'default-'+treekey
	treekey++
	return this
}

Tree.prototype.Add = function(path, data) {
	if (path.length == 0 ||(path.length==1&&path[0]=='')) {
		if (this.childs[data.title] == undefined) {
			this.childs[data.title] = new Tree()
		}
		// this.data.push(data)
		this.childs[data.title].data=data
		this.childs[data.title].key = data.id
	} else {
		if (this.childs[path[0]] == undefined) {
			this.childs[path[0]] = new Tree()
		}
		this.childs[path[0]].Add(path.slice(1), data)
	}
};

Tree.prototype.GetChilds = function() {
	if (  JSON.stringify(this.childs)=='{}') {
		return null
	}
	var mapping = {} // nextid -> id
	var data = []
	var keys =Object.keys(this.childs)
	for (var i of keys) {
		if (this.childs[i].data.nextid == 0) {
			data.push(this.childs[i])
			keys.splice(keys.indexOf(this.childs[i].data.title), 1)
		} else {
			mapping[this.childs[i].data.id] = i
		}
	}
	if (data.length == 0 || data.length==keys.length) {
		return this.childs
	}

	for (var i of data) {
		if(keys.indexOf(i.title)!=-1){
			keys.splice(keys.indexOf(i.title), 1)
		}
	}
	for (;;) {
		var l = data.length
		for (var key of keys) {
			if (this.childs[key].data.nextid == data[0].data.id) {
				data.unshift(this.childs[key])
				keys.splice(keys.indexOf(key), 1)
				continue
			}
			if (this.childs[key].data.id == data[0].data.nextid) {
				data.push(this.childs[key])
				keys.splice(keys.indexOf(key), 1)
				continue
			}
		}
		if (l == data.length) {
			break
		}
	}
	for(var key of keys) {
		data.push(this.childs[key])
	}
	return data
}
Vue.component("Note-Menu-Tree",{
	props:["tree"],
	template:`
	<ul>
		<router-link v-if="tree.data" :to="'/'+$root.getSpaceName()+'/'+(tree.data.directory?tree.data.directory+'/':'')+tree.data.title">{{ tree.data.title }}</router-link>
		<ul class="note-menu-list">
			<Note-Menu-Tree v-for="child in tree.GetChilds()" :key="'note-tree-id-'+child.key":tree="child"/>
		</ul>
	</ul>
	`
})
Vue.component("Note-Menu", {
	data: function() {return {tree: new Tree()} },
	mounted: function() {
		axios.get("spaces/list/" + this.$route.params.username + "/" + this.$route.params.name).then((response) => {
			var tree = new Tree();
			for (var note of response.data||[]) {
				tree.Add(note.directory.split("/"), note);
			}
			this.tree = tree;
			console.log(tree)
		})
	},
	template: `<Note-Menu-Tree :tree="tree"/>`
})
Vue.component("Note-Edit-Text", {
	data: function() {
		return {
			height: "300px", width:"50%",
			isShowPreview: true, isShowText: true,
		}
	},
	props: ["note"],
	mounted: function() {this.resetFormat(); this.resetHight() },
	methods: {
		marked: marked,
		openfile: function(e) {
			var files = this.$refs.markOopenfile.files;
			if(files.length!=1){return }
			var reader = new FileReader();
			reader.onload = (e)=> {this.note.content = e.target.result; };
			reader.readAsText(files[0]);
		},
		showText:function(){
			if(this.isShowText) {this.isShowPreview=true }
			this.isShowText = !this.isShowText;
			this.width = this.isShowText && this.isShowPreview ?"50%":"100%";
		},
		showPreview: function(){
			if(this.isShowPreview) {this.isShowText=true }
			this.isShowPreview = !this.isShowPreview;
			this.width = this.isShowText && this.isShowPreview ?"50%":"100%";
		},
		resetHight: function() {setTimeout(() => {this.height = this.$refs.textEditor.scrollHeight + 'px'; }, 100); },
		resetFormat: function() {this.isShowText=true ; if (this.note.format == 'md') {this.width = "50%"; this.isShowPreview = true; }else {this.width = "100%"; this.isShowPreview = false; } }
	},
	watch: {"note.content": "resetHight","note.format": "resetFormat"},
	template: `
	<div id="note-edit-content">
		<div class="note-edit-md-toolbar">
			<input type="file" ref="markOopenfile" @change="openfile" accept=".html,.txt,.md,.go,.sql,js,css"/>
			<button v-if="note.format=='md'" @click="showText">showmark</button>
			<button v-if="note.format=='md'" @click="showPreview">preview</button>
		</div>
		<div class="note-edit-md-content">
			<div v-if="isShowText" class="note-edit-md-input" :style="{'width': width}"><textarea ref="textEditor" v-model="note.content" :style="{'height': height}"/></div>
			<div v-if="isShowPreview" class="note-edit-md-preview note-content" v-html="marked(note.content)" :style="{'width': width}"></div>
		</div>
	</div>`,
})
Vue.component("Note-Edit-Html", {
	data: function() {
		return {
			editor:null,
			wangEditor: window.wangEditor
		}
	},
	props: ["note"],
	mounted: function() {
		if (!window.wangEditor) {
			const s = document.createElement('script');
			s.type = 'text/javascript';
			s.src = '/js/lib/wangEditor.min.js';
			document.body.appendChild(s);
			var loadtimer = self.setInterval(() => {
				if (window.wangEditor) {
					loadtimer = window.clearInterval(loadtimer)
					this.wangEditor = window.wangEditor
					this.initeditor()
				}
			}, 100);
			return
		}
		this.initeditor()
	},
	methods:{
		initeditor: function() {
			this.editor = new wangEditor(this.$refs.htmlEditor)
  			this.editor.customConfig.zIndex = 0
			this.editor.customConfig.onchange = (html) => {this.note.content = html }
			this.editor.create()
			this.editor.txt.html(this.note.content)
		}
	},
	template: `<div id="note-edit-content">
		<div class="note-edit-html-content" ref="htmlEditor">
		</div>
	</div>`
})

var routes = {
	path: "*",
	mounted: function() {
		if(this.$root.Base.name != "") {
			window.location.hash="#/spaces/"+this.$root.Base.name;
		}else {
			window.location.hash="#/spaces/";
		}
	},
	template: `<div></div>`,
	Index: {
		template: "<div>User {{ $route.params.id }}</div>"
	},
	Spaces: {
		My: {
			path: "/:key",
			data: function() {
				return {
					spaces: [],
				}
			},
			mounted: function() {
				this.init()
			},
			methods: {
				init: function() {
					axios.get("spaces/info/" + this.$route.params.key).then((response) => {
						this.spaces = response.data
					})
				}
			},
			template: `<div>
	<nav class="box">
		<router-link :to="'/spaces/'+$root.Base.name" class="button">My Spaces</router-link>
		<router-link to="/spaces/" class="button">Public Spaces</router-link>
		<router-link to="/content/" class="button" :style="{'display':'none'}">Public Note</router-link>
	</nav>
	<div>
		<div v-for="space in spaces" class="note-spaces">
			<router-link :to="'/spaces/'+$root.Base.name+'/'+space.name">{{ space.name }}</router-link>
		</div>
		<div class="note-spaces">
			<router-link to="/new/spaces">new spaces</router-link>
		</div>
	</div>
</div>`
		},
		Public: {
			path: "/",			
			data: function() {return {spaces: [], } },
			mounted: function() {this.init() },
			methods: {
				init: function() {
					axios.get("spaces/public").then((response) => {
						this.spaces = response.data
					})
				}
			},
			template: `<div>
	<nav class="box">
		<router-link :to="'/spaces/'+$root.Base.name" class="button">My Spaces</router-link>
		<router-link to="/spaces/" class="button">Public Spaces</router-link>
	</nav>
	<div>
		<div v-for="space in spaces" class="note-spaces">
			<router-link :to="'/spaces/'+$root.Base.name+'/'+space.name">{{ space.name }}</router-link>
		</div>
	</div>
</div>`
		},
		Spaces: {
			path: "/:username/:name",
			data: function() {
				return {space:{} }
			},
			mounted: function() {
				this.init()
			},
			methods: {
				init: function() {
					var params = this.$route.params
					axios.get(`spaces/info/${params.username}/${params.name}`).then((response)=>{
						console.log(response.data)
						this.space = response.data
					})
				}
			},
			template: `
	<div id="note-container">
		<div id="note-tree">
			<Note-Menu/>
		</div>
		<div>
			spaces setting
			<p>name: {{ space.name }}</p>
			<p>public: {{ space.public }}</p>
			<button>delete</button>
			<router-link :to="'/new/'+$route.params.username+'/'+$route.params.name">new note</router-link>
		</div>
	</div>`,
			Note: {
				path: "/*",
				data: function() {
					return {
						note: {
							content: "",
						},
					}
				},
				mounted: function() {
					this.init()
				},
				methods: {
					init: function() {
						axios.get("content"+this.$route.path.slice(7)).then((response) => {
							this.note = response.data
							document.body.scrollTop = document.documentElement.scrollTop = 0;
						})
					},
				},
				computed: {
					getNoteContent: function() {
						if (this.note.format == 'md') {
							return marked(this.note.content)
						} else if (this.note.format == 'url-md') {
							return 'server not get resource: <a target="view_window" href=' + this.note.content + '>' + this.note.content + '</a>'
						} else if (this.note.format == 'html') {
							var dom = document.createElement("div")
							dom.innerHTML = this.note.content
							for(var i of dom.querySelectorAll(`code[class*="language-"],pre[class*="language-"]`)){
								var lang =i.className.match(/language-(\w+)/)
								i.innerHTML = Prism.highlight(i.innerText, Prism.languages[lang[1]])
							}
							return filterXSS(dom.innerHTML)
						}
						return this.note.content
					},
					getNoteDate() {
						var notedate = new Date(this.note.edittime);
						var today = new Date();
						if(notedate.setHours(0,0,0,0)==today.setHours(0,0,0,0)){
							return notedate.Format("yyyy-MM-dd hh:mm")
						} 
						return notedate.Format("yyyy-MM-dd")
					}
				},
				watch: {
					"$route": "init"
				},
				template: `
	<div id="note-container">
		<div id="note-tree">
			<Note-Menu/>
		</div>
		<div id="note-page" :key="note.directory+'/'+note.name">
			<div class="note-user">
			</div>
			<div class="note-header">
				<div class="row flex-between">
					<h1>{{note.title}}</h1>
					<div>
						<button>delete</button>
						<router-link class="button" :to="'/edit'+this.$route.path.slice(7)">edit</router-link>
						<router-link class="button" :to="'/new'+this.$route.path.slice(7)">new</router-link>
					</div>
				</div>
				<span class="time">修改时间 {{ getNoteDate }}</span>
				<hr/>
			</div>
			<div class="note-content" v-html="getNoteContent"></div>
			<div class="note-comment"></div>
		</div>
	</div>`,
			}
		},
	},
	Edit: {
		Spaces: {
			path: "/:username/:name",
			template: `222`,
			Note: {
				path: "/*",
				data: function() {
					return {
						note: {
							id:0,
							title:"",
							format:"",
							content: "",
							tags:[],
						},
					}
				},
				mounted: function() {
					this.init()
				},
				methods: {
					init: function() {
						console.log(this)
						console.log(this.$router)
						axios.get('content'+this.$route.path.slice(5),{
							params:{parseformat:false, }
						}).then((response) => {
							this.note = response.data
							if(this.note.tags==null){this.note.tags=[]} 
							this.note.path = (this.note.directory?this.note.directory+"/":"") + this.note.title
							document.body.scrollTop = document.documentElement.scrollTop = 0;
						})
					},
					inputtags: function(e) {
						if (e.type == "keydown" && e.keyCode != 9) {return }
						var tag = this.$refs.inputtags.value;
						this.$refs.inputtags.value = "";
						if (tag == "") {return }
						if (e.type == "keydown" && e.keyCode == 9) {e.preventDefault() }
						for (var t of this.note.tags) {if (t == tag) {return } }
						this.note.tags.push(tag);
					},
					remotetag: function(tag){this.note.tags.remove(tag)},
					commit: function(){
						var params = this.$route.params
						var content = {title:this.note.title, directory:this.note.directory, format: this.note.format, tags: this.note.tags, content: this.note.content}
						axios.post(`content/${params.username}/${params.name}/${params.pathMatch}`, content).then(response => {
							console.log(response)
							window.location.hash=`#/spaces/${params.username}/${params.name}/${params.pathMatch}`
						})
					}
				},
				template: `<div id="note-edit">
		<div id="edit-header">
			<div class="row flex-between">
				<div>
					<select v-model="note.format">
						<option v-for="format in $root.NoteFormat" :key="format" :value="format">{{format}}</option>
					</select>
					<input v-model="note.path">
				</div>
				<button @click="commit">commit</button>
			</div>
			<div>
				<a v-for="tag in note.tags">{{tag}}<button @click="remotetag(tag)">x</button></a>
				<input ref="inputtags" @keydown="inputtags" @blur="inputtags" placeholder="tags"/>
			</div>
		</div>
		<Note-Edit-Html v-if="note.format=='html'" :note="note" />
		<Note-Edit-Text v-else-if="note.format!=''" :note="note" />

	</div>`,
			},
		},
	},
	New: {
		Spaces: {
			data: function() {
				return {
					value: {},
				}
			},
			methods: {
				commit: function(){
					console.log(this.value)
					axios.put("spaces/new", {
						name: this.value.name,
						public: this.value.public,
					})
				},
			},
			template: `<div>new spaces			
		<table>
			<tbody>
				<tr>
					<td>name</td>
					<td><input v-model="value.name"></input></td>
				</tr>
				<tr>
					<td>public</td>
					<td><input type="checkbox" name="public" v-model="value.public"/></td>
				</tr>
			</tbody>
		</table>
		<button @click="commit">commit</button>
		<button @click="$root.$router.go(-1)">cannel</button>
			</div>`
		},
		Note: {
			path: "/:username/:name",
			data: function() {
				return {
					path: ""
				}
			},
			methods: {
				commit: function() {
					var params = this.$route.params
					axios.put(`content/${params.username}/${params.name}/${this.path}`, {format:"md"}).then(()=>{
						window.location.hash=`#/edit/${params.username}/${params.name}/${this.path}`
					})
				}
			},
			template:`<div>
				note name: <input v-model="path"></input>
				<button @click="commit">create note</button>
			</div>`
		},
	}
}

console.log( getRoutes("", routes))

var options = filterXSS.whiteList 
options["pre"]=['class']
options["code"]=['class']
options["span"]=['class']
var vm = new Vue({
	el: '#app',
	template: `<app/>`,
	router: new VueRouter({routes: getRoutes("", routes)}),
	data: function() {
		return {
			Base: Base,
			NoteFormat:["md","text","html","url-md"],
			getRouteUrl: function(path) {return path.replace(/(\$\w+)/g, (arg1) => {return this.$route.params[arg1.slice(1)] }) },
			getRouteKey: function() {var key = this.$route.params.key; if (/^\d+$/.test(key)) {return 'id/' + key } return 'name/' + key },
			getApiVersion: function() {return apiVersion },
			getSpaceName: function(){
				return "spaces/" + this.$route.params.username + "/" + this.$route.params.name
			}
		}
	},
})
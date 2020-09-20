//

apiVersion = "/api/v1/term"
var routes = {
	path: "*",
	template: "<div>User index {{ $route.params.id }}</div>",
	Index: {
		template: "<div>User {{ $route.params.id }}</div>"
	},
	User: {
		Index: {
			template: `
			<div id="nav-content" class="box" >
				<span>User</span>
				<TableController-Index name="user">
					<template v-slot:id="props">
						<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
					</template>
					<template v-slot:name="props">
						<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
					</template>
				</TableController-Index>
			</div>
			`
		},
		New: {
			template: `<TableController-New id="nav-content" class="box" name="user"/>`
		},
		Info: {
			path: "/:key",

			data: function() {
				return {
					extraFields: ['options'],
					currentTab: "Auth info",
					tabs: ["Auth info", "User all host", "User all group"]
				}
			},
			computed: {
				dataUserhostUrl: function() {
					var key = this.$route.params.key
					if (/^\d+$/.test(key)) {
						return '/api/v1/term/userhost/view/search/userid/' + key
					}
					return '/api/v1/term/userhost/view/search/username/' + key
				},
				dataUserhostgroupUrl: function() {
					var key = this.$route.params.key
					if (/^\d+$/.test(key)) {
						return '/api/v1/term/userhostgroup/view/search/userid/' + key
					}
					return '/api/v1/term/userhostgroup/view/search/username/' + key
				},
			},
			template: `
			<div id="nav-content" class="box">
				<div>Term > User > Info </div>
				<div>
					<img src="/api/v1/auth/user/icon/id/2" class="avatar">
					<div>Name</div>
					<div>Name2</div>
				</div>
				<TableController-Info name="user">
					<template v-slot:info-bar>
						<button v-for="tab in tabs" v-bind:key="tab" v-on:click="currentTab = tab">{{ tab }}</button>
						<TableController-Data v-if="currentTab=='User all host'" :tableFields="$root.fields.userhostHost" :extraFields="extraFields" :url="dataUserhostUrl">
							<template v-slot:userid="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:username="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:options="props">
								<button>remove</button>
							</template>
						</TableController-Data>
						<TableController-Data v-if="currentTab=='User all group'" :tableFields="$root.fields.userhostHostgroup" :extraFields="extraFields" :url="dataUserhostgroupUrl">
							<template v-slot:userid="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:username="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:options="props">
								<button>remove</button>
							</template>
						</TableController-Data>
					</template>
				</TableController-Info>
			</div>`

		},
	},
	Host: {
		Index: {
			template: ` 
			<TableController-Index id="nav-content" class="box" name="host" :extraFields="['connect']">
				<template v-slot:id="props">
					<router-link :to="'/host/info/'+props.value" class="button">{{ props.value }}</router-link>
				</template>
				<template v-slot:name="props">
					<router-link :to="'/host/info/'+props.value" class="button">{{ props.value }}</router-link>
				</template>
				<template v-slot:connect="props">
					<router-link :to="'/host/connect/'+props.data.id" class="button">connect</router-link>
				</template>
			</TableController-Index>				 
		`
		},
		New: {
			template: `<TableController-New id="nav-content" class="box" name="host"/>`
		},
		Info: {
			path: "/:key",
			data: function() {
				return {
					currentTab: "Home",
					tabs: ["Home", "Posts", "Archive"]
				}
			},
			computed: {
				dataurl: function() {
					var key = this.$route.params.key
					if (/^\d+$/.test(key)) {
						return '/api/v1/term/userhost/view/search/userid/' + key
					}
					return '/api/v1/term/userhost/view/search/username/' + key
				}
			},
			template: `<TableController-Info id="nav-content" class="box" name="host">
				<template v-slot:info-bar>
					<button v-for="tab in tabs" v-bind:key="tab">{{ tab }}</button>
					<TableController-Data :tableFields="$root.fields.userhostUser" :url="dataurl">
						<template v-slot:userid="props">
							<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
						</template>
						<template v-slot:username="props">
							<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
						</template>
					</TableController-Data>
				</template>
			</TableController-Info>`
		},
		Connect: {
			path: "/connect/:key",
			data: function() {
				return {
					options: {}
				}
			},
			mounted: function() {
				// Object.keys(props).forEach(key => this.$set(this.options, key, this[key]))
				let term = new Terminal(this.options)
				const fitAddon = new FitAddon();
				term.open(document.getElementById('terminal'));
				term.loadAddon(fitAddon);
				this.$terminal = term
				this.$fit = fitAddon
				this.$fit.fit()



				let ws = new WebSocket(this.getUrl());
				this.$ws = ws
				ws.binaryType = "blob";
				ws.onopen = () => {
					term.write('\r\nWelcome to eudore web shell !\r\n\r\n')
					this.sendMessage({
						type: "env",
						message: {
							name: "SSS",
							value: "sss",
						},
					})
					this.sendMessage({
						type: "pty-req",
						message: {
							term: "xterm",
							columns: term.cols,
							rows: term.rows,
						},
					})
					this.sendMessage({
						type: "shell",
					})
				}
				ws.onmessage = function(event) {
					if (event.data instanceof Blob) {
						var reader = new FileReader();
						reader.onload = function() {
							term.write(reader.result)
						}
						reader.readAsText(event.data);
					} else {
						var response = JSON.parse(event.data)
						console.log(response)
						switch (response.type) {
							case "start":
								// vnode.state.timer = setInterval(function() {
								// 	ws.send(JSON.stringify({
								// 		type: "ping",
								// 	}))
								// }, 20000)
						}
					}
				};


				term.onData(function(data) {
					ws.send(new Blob([data]))
				})
				term.onResize(size => {
					this.sendMessage({
						type: "window-change",
						message: {
							columns: size.cols,
							rows: size.rows,
						},
					})
				})
				window.addEventListener("resize", (event) => {
					this.$fit.fit()
				})
			},
			beforeDestroy: function() {
				this.$terminal.selectAll()
				this.$emit('update:buffer', this.$terminal.getSelection().trim())
				this.$terminal.dispose()
				this.$ws.close()
			},
			methods: {
				getUrl: function(vnode) {
					var connnetParams = new URL(window.location.href).searchParams || new URLSearchParams()
					connnetParams.append("hostid", this.$route.params.key)
					if (Base.bearer != "") {
						connnetParams.append("bearer", Base.bearer)
					}
					if (window.location.protocol == "http:") {
						return "ws://" + location.host + "/api/v1/term/connect?" + connnetParams.toString();
					}
					return "wss://" + location.host + "/api/v1/term/connect?" + connnetParams.toString();
				},
				sendData: function(data) {
					this.$ws.send(new Blob[data])
				},
				sendMessage: function(msg) {
					this.$ws.send(JSON.stringify(msg))
				},
			},
			template: `<div id="nav-content" class="box" style="width: 100%;"><div id="terminal"/></div>`,
		},
	},
	Hostgroup: {
		Index: {
			template: `<TableController-Index id="nav-content" class="box" name="hostgroup">
			<template v-slot:id="props">
				<router-link :to="'/hostgroup/info/'+props.value" class="button">{{ props.value }}</router-link>
			</template>
			<template v-slot:name="props">
				<router-link :to="'/hostgroup/info/'+props.value" class="button">{{ props.value }}</router-link>
			</template>
		</TableController-Index>`
		},
		New: {
			template: `<TableController-New id="nav-content" class="box" name="hostgroup"/>`
		},
		Info: {
			path: "/:key",
			template: `<TableController-Info id="nav-content" class="box" name="hostgroup"/>`
		},
	},
	Video: {
		Index: {
			template: `<TableController-Index id="nav-content" class="box" name="video" :extraFields="['open']">
			<template v-slot:name="props">
				<router-link :to="'/video/info/'+props.value" class="button">{{ props.value }}</router-link>
			</template>
			<template v-slot:startstamp="props">
				<template>{{ (new Date(props.value )).Format("yyyy-MM-dd hh:mm") }}</template>
			</template>
			<template v-slot:endstamp="props">
				<template>{{ (new Date(props.value )).Format("yyyy-MM-dd hh:mm") }}</template>
			</template>
			<template v-slot:open="props">
				<router-link :to="'/video/open/'+props.data.name" class="button">open</router-link>
			</template>
		</TableController-Index>`
		},
		Info: {
			path: "/:key",
			template: `<TableController-Info id="nav-content" class="box" name="video"/>`
		},
		Open: {},
	}
}

var vm = new Vue({
	el: '#app',
	template: `<app/>`,
	router: new VueRouter({
		routes: getRoutes("", routes),
	}),
	data: function() {
		return {
			Base: Base,
			navs: ["User", "Host", "HostGroup", "Video", "Logger"],
			fields: {
				user: ["id", "name", "password", "salf", "publickey"],
				host: ["id", "name", "protocol", "addr", "user", "password", "privateky"],
				hostgroup: ["id", "name", "description"],
				video: ["name", "user", "remoteaddr", "localaddr", "startstamp", "endstamp", "savedir"],
				userhostUser: ["userid", "username", "granttime"],
				userhostHost: ["hostid", "hostname", "protocol", "addr", "user", "config", "granttime"],
				userhostHostgroup: ["userid", "username", "password", "salf", "publickey", "groupid", "groupname", "groupdescription", "granttime"],
			},
		}
	},
})
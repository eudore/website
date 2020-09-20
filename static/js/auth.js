//


apiVersion = "/api/v1/auth"
var statuslist = ["正常", "冻结", "禁用"]

function statusstr(status) {
	return statuslist[status]
}

function getStringKey(key) {
	if (/^\d+$/.test(key)) {
		return 'id'
	}
	return 'name'
}

var routes = {
	path: "*",
	template: "<div>User index {{ $route.params.id }}</div>",
	Index: {
		template: "<div>User {{ $route.params.id }}</div>"
	},
	User: {
		Index: {
			template: `<div id="nav-content" class="box"><TableController-Index name="user"/></div>` },
		New: {
			template: `<TableController-New id="nav-content" class="box" name="user"/>`,
		},
		Info: {
			path: "/info/:key",
			data: function() {
				return {
					data: {}, extraFields: ['options'], currentTab: "User Info",
					tabs: ["User Info", "User Permission", "User Role", "User Policy"]
				}
			},
			mounted: function() {this.init() },
			methods: {
				init: function() {axios.get(apiVersion + "/user/" + this.$root.getRouteKey()).then((response) => {this.data = response.data})},
				removeuserpermission: function(data) {axios.delete('userpermission/', {data: {userid: data.userid,permissionid: data.permissionid}})},
				removeuserrole: function(data) {axios.delete('userrole/', {data: {userid: data.userid,roleid: data.roleid}})},
				removeuserpolicy: function(data) {axios.delete('userpolicy/', {data: {userid: data.userid,policyid: data.policyid}})},
			},
			template: `
			<div id="nav-content" class="box">
				<div>Auth > User > Info </div>
				<div>
					<img v-if="data.id" :src="$root.getApiVersion()+'/icon/'+data.id" class="avatar">
					<div>{{ data.name }}</div>
					<div>Name2</div>
				</div>
				<div>
					<button v-for="tab in tabs" v-bind:key="'nav'+tab" v-on:click="currentTab = tab">{{ tab }}</button>
				</div>
				<div>				
					<TableController-Info v-if="currentTab=='User Info'" :key="currentTab" name="user"/>

					<template v-if="currentTab=='User Permission'"> 
						<router-link :to="'/user/bind/'+data.id+'/permission'" class="button">Bind Permission</router-link>
						<TableController-Data :key="currentTab" :tableFields="$root.fields.userpermission" 
						:extraFields="extraFields" :url="'userpermission/view/search/user'+this.$root.getRouteKey()">
							<template v-slot:userid="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:username="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:options="props">
								<button @click="removeuserpermission(props.data)">remove</button>
							</template>
						</TableController-Data>
					</template>

					<template v-if="currentTab=='User Role'"> 
						<router-link :to="'/user/bind/'+data.id+'/role'" class="button">Bind Role</router-link>
						<TableController-Data :key="currentTab" :tableFields="$root.fields.userrole" 
						:extraFields="extraFields" :url="'userrole/view/search/user'+this.$root.getRouteKey()">
							<template v-slot:userid="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:username="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:options="props">
								<button @click="removeuserrole(props.data)">remove</button>
							</template>
						</TableController-Data>
					</template>

					<template v-if="currentTab=='User Policy'"> 
						<router-link :to="'/user/bind/'+data.id+'/policy'" class="button">Bind Policy</router-link>
						<TableController-Data :key="currentTab" :tableFields="$root.fields.userpolicy" 
						:extraFields="extraFields" :url="'userpolicy/view/search/user'+this.$root.getRouteKey()">
							<template v-slot:userid="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:username="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:options="props">
								<button @click="removeuserpolicy(props.data)">remove</button>
							</template>
						</TableController-Data>
					</template>
				</div>
			</div>`
		},
		Bind: {
			path: "/bind/:key",
			Permission: {
				template: `<TableController-Bind source="user" target="permission" listurl='permission/list' haveurl='userpermission/view/search/userid/$key'
				commiturl='userpermission/news' :grantFields='$root.fields.grantallow' :tableFields='$root.fields.grantcannel'/>`,
			},
			Role: {
				template: `<TableController-Bind source="user" target="role" listurl='role/list' haveurl='userrole/view/search/userid/$key'
				commiturl='userrole/news' :grantFields='$root.fields.grantdefault' :tableFields='$root.fields.grantcannel'/>`,
			},
			Policy: {
				template: `<TableController-Bind source="user" target="policy" listurl='policy/list' haveurl='userpolicy/view/search/userid/$key'
				commiturl='userpolicy/news' :grantFields='$root.fields.grantdefault' :tableFields='$root.fields.grantcannel'/>`,
			},
		},
	},
	Permission: {
		Index: {
			template: `<div id="nav-content" class="box"><TableController-Index name="permission"/></div> `
		},
		New: {
			template: `<TableController-New id="nav-content" class="box" name="permission"/>`
		},
		Info: {
			path: "/info/:key",
			data: function() {
				return {
					data: {},
					extraFields: ['options'],
					currentTab: "Permission Info",
					tabs: ["Permission Info", "Permission User", "Permission Role"]
				}
			},
			mounted: function() {
				this.init()
			},
			methods: {
				init: function() {
					axios.get(apiVersion + "/permission/" + this.$root.getRouteKey()).then((response) => {
						this.data = response.data
					})
				},
				removeuserpermission: function(data) {axios.delete('userpermission/', {data: {permissionid: data.permissionid,userid: data.userid}})},
				removerolepermission: function(data) {axios.delete('rolepermission/', {data: {permissionid: data.permissionid,roleid: data.roleid}})},
			},
			template: `
			<div id="nav-content" class="box">
				<div>Auth > Permission > Info </div>
				<div>
					<div>{{ data.name }}</div>
				</div>
				<div>				
					<button v-for="tab in tabs" v-bind:key="'nav'+tab" v-on:click="currentTab = tab">{{ tab }}</button>
					<TableController-Info v-if="currentTab=='Permission Info'" :key="currentTab" name="permission"/>

					<template v-if="currentTab=='Permission User'">
						<router-link :to="'/permission/bind/'+data.id+'/user'" class="button">Bind User</router-link>
						<TableController-Data :key="currentTab" :tableFields="$root.fields.userpermission" 
						:extraFields="extraFields" :url="'userpermission/view/search/permission'+this.$root.getRouteKey()">
							<template v-slot:userid="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:username="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:options="props">
								<button @click="removeuserpermission(props.data)">remove</button>
							</template>
						</TableController-Data>
					</template>

					<template v-if="currentTab=='Permission Role'">
						<router-link :to="'/permission/bind/'+data.id+'/role'" class="button">Bind Role</router-link>
						<TableController-Data :key="currentTab" :tableFields="$root.fields.rolepermission" 
						:extraFields="extraFields" :url="'rolepermission/view/search/permission'+this.$root.getRouteKey()">
							<template v-slot:userid="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:username="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:options="props">
								<button @click="removerolepermission(props.data)">remove</button>
							</template>
						</TableController-Data>
					</template>
				</div>
			</div>`
		},
		Bind: {
			path: "/bind/:key",
			User: {
				template: `<TableController-Bind source="permission" target="user" listurl='user/list' haveurl='userpermission/view/search/permissionid/$key'
				commiturl='userpermission/news' :grantFields='$root.fields.grantallow' :tableFields='$root.fields.grantcannel'/>`,
			},
			Role: {
				template: `<TableController-Bind source="permission" target="role" listurl='role/list' haveurl='rolepermission/view/search/permissionid/$key'
				commiturl='rolepermission/news' :grantFields='$root.fields.grantdefault' :tableFields='$root.fields.grantcannel'/>`,
			},
		},
	},
	Role: {
		Index: {
			template: `
			<div id="nav-content" class="box">
				<TableController-Index name="role"> 
				</TableController-Index>
			</div>
			`
		},
		New: {
			template: `<TableController-New id="nav-content" class="box" name="role"/>`
		},
		Info: {
			path: "/info/:key",
			data: function() {
				return {
					data: {},
					extraFields: ['options'],
					currentTab: "Role Info",
					tabs: ["Role Info", "Role User", "Role Permission"]
				}
			},
			mounted: function() {
				this.init()
			},
			methods: {
				init: function() {
					axios.get(apiVersion + "/role/" + this.$root.getRouteKey()).then((response) => {
						this.data = response.data
					})
				},
				removeuser: function(data) {axios.delete('userrole/', {data: {roleid: data.roleid,userid: data.userid}})},
				removepermission: function(data) {axios.delete('rolepermission/', {data: {roleid: data.roleid,permissionid: data.permissionid}})},
			},
			template: `
			<div id="nav-content" class="box">
				<div>Auth > Role > Info </div>
				<div>
					<div>{{ data.name }}</div>
				</div>
				<div>				
					<button v-for="tab in tabs" v-bind:key="'nav'+tab" v-on:click="currentTab = tab">{{ tab }}</button>
					<TableController-Info v-if="currentTab=='Role Info'" :key="currentTab" name="role"/>

					<template v-if="currentTab=='Role User'">
						<router-link :to="'/role/bind/'+data.id+'/user'" class="button">Bind User</router-link>
						<TableController-Data :key="currentTab" :tableFields="$root.fields.userrole" 
						:extraFields="extraFields" :url="'userrole/view/search/role'+this.$root.getRouteKey()">
							<template v-slot:userid="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:username="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:options="props">
								<button @click="removeuser(props.data)">remove</button>
							</template>
						</TableController-Data> 
					</template>

					<template v-if="currentTab=='Role Permission'">
						<router-link :to="'/role/bind/'+data.id+'/permission'" class="button">Bind Permission</router-link>
						<TableController-Data :key="currentTab" :tableFields="$root.fields.rolepermission" 
						:extraFields="extraFields" :url="'rolepermission/view/search/role'+this.$root.getRouteKey()">
							<template v-slot:userid="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:username="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:options="props">
								<button @click="removepermission(props.data)">remove</button>
							</template>
						</TableController-Data> 
					</template>
				</div>
			</div>`
		},
		Bind: {
			path: "/bind/:key",
			User: {
				template: `<TableController-Bind source="role" target="user" listurl='user/list' haveurl='userrole/view/search/userid/$key'
				commiturl='userrole/news' :grantFields='$root.fields.grantdefault' :tableFields='$root.fields.grantcannel'/>`,
			},
			Permission: {
				template: `<TableController-Bind source="role" target="permission" listurl='permission/list' haveurl='rolepermission/view/search/roleid/$key'
				commiturl='rolepermission/news' :grantFields='$root.fields.grantdefault' :tableFields='$root.fields.grantcannel'/>`,
			},
		},
	},
	Policy: {
		Index: {
			template: `
			<div id="nav-content" class="box">
				<TableController-Index name="policy">
				</TableController-Index>
			</div>
			`
		},
		New: {
			template: `<TableController-New id="nav-content" class="box" name="policy"/>`
		},
		Info: {
			path: "/info/:key",
			data: function() {
				return {
					data: {},
					extraFields: ['options'],
					currentTab: "Policy Content",
					tabs: ["Policy Content", "Policy Version","Policy User"]
				}
			},
			mounted: function() {
				this.init()
			},
			methods: {
				init: function() {
					axios.get(apiVersion + "/policy/" + this.$root.getRouteKey()).then((response) => {
						this.data = response.data
					})
				},
				removeuser: function(data) {axios.delete('userpolicy/', {data: {userid: data.userid,policyid:data.policyid}})},
			},
			template: `
			<div id="nav-content" class="box">
				<div>Auth > Policy > Info </div>
				<div>
					<div>{{ data.id }}</div>
					<div>{{ data.name }}</div>
					<div>{{ data.version }}</div>
					<div>{{ data.description }}</div>
					<div>{{ data.time }}</div>
				</div>
				<div>				
					<button v-for="tab in tabs" v-bind:key="'nav'+tab" v-on:click="currentTab = tab">{{ tab }}</button>
					<div v-if="currentTab=='Policy Content'">
						<button>edit policy</button>
						<pre>{{JSON.stringify(JSON.parse(data.policy||'{}'), null, '\t')}}</pre>
					</div>

					<TableController-Data v-if="currentTab=='Policy Version'" :key="currentTab" :tableFields="$root.fields.policy" :extraFields="extraFields" :url="'policyversion/id/'+data.id">
						<template v-slot:userid="props">
							<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
						</template>
						<template v-slot:username="props">
							<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
						</template>
						<template v-slot:options="props">
							<button @click="removeuser(props.data)">remove</button>
						</template>
					</TableController-Data> 

					<template v-if="currentTab=='Policy User'">
						<router-link :to="'/policy/bind/'+data.id+'/user'" class="button">Bind User</router-link>
						<TableController-Data :key="currentTab" :tableFields="$root.fields.userpolicy" :extraFields="extraFields" :url="'userpolicy/view/search/policy'+this.$root.getRouteKey()">
							<template v-slot:userid="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:username="props">
								<router-link :to="'/user/info/'+props.value" class="button">{{ props.value }}</router-link>
							</template>
							<template v-slot:options="props">
								<button @click="removeuser(props.data)">remove</button>
							</template>
						</TableController-Data> 
					</template>
				</div>
			</div>`
		},
		Bind: {
			path: "/bind/:key",
			User: {
				template: `<TableController-Bind source="policy" target="user" listurl='user/list' haveurl='userpolicy/view/search/userid/$key'
				commiturl='userpolicy/news' :grantFields='$root.fields.grantdefault' :tableFields='$root.fields.grantcannel'/>`,
			},
		}
	}
}

var vm = new Vue({
	el: '#app',
	template: `<app/>`,
	router: new VueRouter({routes: getRoutes("", routes), }),
	data: function() {
		return {
			Base: Base,
			navs: ["User", "Permission", "Role", "Policy"],
			fields: {
				user: ["id", "name", "status", "level", "mail", "tel", "loginip", "logintime", "sigintime", "lang"],
				permission: ["id", "name", "description", "time"],
				role: ["id", "name", "description", "time"],
				policy: ["id", "name", "description", "policy", "time"],

				userpermission: ["userid", "permissionid", "username", "permissionname", "effect", "description", "granttime"],
				userrole: ["userid", "roleid", "username", "rolename", "description", "granttime"],
				userpolicy: ["userid", "policyid", "username", "policyname", "description", "granttime"],
				rolepermission: ["permissionid", "permissionname", "roleid", "rolename", "granttime"],

grantdefault:["id","name","grant"],
grantcannel:["id","name","cannel"],
grantallow:["id", "name", 'allow', 'deny'],
grantuserrole:[],
grantuserpolicy:[],

grantpermissionuser:["id","name","allow","deny"],
				grantpermission: ["id", "name", "description", "time", 'allow', 'deny'],
				grantpolicy: ["id", "name", "description", "time"],
			},
			getRouteUrl: function(path) {return path.replace(/(\$\w+)/g, (arg1) => {return this.$route.params[arg1.slice(1)] }) },
			getRouteKey: function() {var key = this.$route.params.key; if (/^\d+$/.test(key)) {return 'id/' + key } return 'name/' + key },
			getApiVersion: function() {return apiVersion },
		}
	},
})
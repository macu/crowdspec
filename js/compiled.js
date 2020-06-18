!function(e){var t={};function n(o){if(t[o])return t[o].exports;var s=t[o]={i:o,l:!1,exports:{}};return e[o].call(s.exports,s,s.exports,n),s.l=!0,s.exports}n.m=e,n.c=t,n.d=function(e,t,o){n.o(e,t)||Object.defineProperty(e,t,{enumerable:!0,get:o})},n.r=function(e){"undefined"!=typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(e,"__esModule",{value:!0})},n.t=function(e,t){if(1&t&&(e=n(e)),8&t)return e;if(4&t&&"object"==typeof e&&e&&e.__esModule)return e;var o=Object.create(null);if(n.r(o),Object.defineProperty(o,"default",{enumerable:!0,value:e}),2&t&&"string"!=typeof e)for(var s in e)n.d(o,s,function(t){return e[t]}.bind(null,s));return o},n.n=function(e){var t=e&&e.__esModule?function(){return e.default}:function(){return e};return n.d(t,"a",t),t},n.o=function(e,t){return Object.prototype.hasOwnProperty.call(e,t)},n.p="",n(n.s=11)}([function(e,t){e.exports=jQuery},function(e,t){e.exports=Vue},function(e,t,n){},function(e,t,n){},function(e,t,n){},function(e,t,n){},function(e,t,n){},function(e,t,n){},function(e,t,n){},function(e,t){e.exports=Vuex},function(e,t){e.exports=VueRouter},function(e,t,n){e.exports=n(20)},function(e,t,n){},function(e,t,n){"use strict";var o=n(2);n.n(o).a},function(e,t,n){"use strict";var o=n(3);n.n(o).a},function(e,t,n){"use strict";var o=n(4);n.n(o).a},function(e,t,n){"use strict";var o=n(5);n.n(o).a},function(e,t,n){"use strict";var o=n(6);n.n(o).a},function(e,t,n){"use strict";var o=n(7);n.n(o).a},function(e,t,n){"use strict";var o=n(8);n.n(o).a},function(e,t,n){"use strict";n.r(t);n(12);var o=n(1),s=n.n(o),i=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticClass:"app"},[n("div",{staticClass:"header"},[n("h1",{on:{click:function(t){return e.gotoIndex()}}},[e._v("CrowdSpec")]),e._v(" "),n("span",[e._v(e._s(e.username))]),e._v(" "),n("button",{attrs:{size:"mini"},on:{click:function(t){return e.logout()}}},[e._v("Log out")])]),e._v(" "),n("div",{staticClass:"content-area"},[n("router-view")],1)])};i._withStripped=!0;var r=n(0),c=n.n(r),a=n(9),u=n.n(a);s.a.use(u.a);var l=c()(window),p=new u.a.Store({state:{windowWidth:l.width()},getters:{userID:function(e){return window.user.id},username:function(e){return window.user.username},dialogTinyWidth:function(e){return e.windowWidth<=767?"80%":e.windowWidth<=991?"50%":"30%"},dialogSmallWidth:function(e){return e.windowWidth<=767?"90%":e.windowWidth<=991?"75%":"50%"},dialogLargeWidth:function(e){return e.windowWidth<=991?"95%":"90%"}},mutations:{setWindowWidth:function(e,t){e.windowWidth=t}}}),d=p;l.on("resize",(function(){p.commit("setWindowWidth",l.width())}));var f=n(10),v=n.n(f),h=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticClass:"index-page"},[n("el-button",{on:{click:function(t){return e.gotoNewSpec()}}},[e._v("New spec")]),e._v(" "),n("div",{staticClass:"user-specs"},[n("h2",[e._v("Your specs")]),e._v(" "),e.loading?n("p",[e._v("Loading...")]):e.userSpecs&&e.userSpecs.length?n("ul",e._l(e.userSpecs,(function(t){return n("li",{key:t.id},[n("router-link",{attrs:{to:{name:"spec",params:{specId:t.id}}}},[e._v(e._s(t.name))])],1)})),0):n("p",[e._v("You do not have any specs.")])])],1)};function b(e){if(console.error(e),e){var t=null;if(0===e.readyState?t="Could not connect to server":e.readyState&&e.status?t="Request failed with error code "+e.status:e.message?t=e.message:"string"==typeof e&&(t=e),t)return void s.a.prototype.$alert(t,"Error",{confirmButtonText:"Ok",type:"error"})}s.a.prototype.$alert("An error occurred",{confirmButtonText:"Ok",type:"error"})}h._withStripped=!0;var _={data:function(){return{userSpecs:[],loading:!0}},mounted:function(){this.reloadSpecs()},beforeRouteUpdate:function(e,t,n){this.reloadSpecs(),n()},methods:{reloadSpecs:function(){var e=this;this.loading=!0,c.a.get("/ajax/user-specs").then((function(t){e.userSpecs=t,e.loading=!1})).fail((function(t){e.loading=!1,b(t)}))},gotoNewSpec:function(){this.$router.push({name:"new-spec"})}}};n(13);function m(e,t,n,o,s,i,r,c){var a,u="function"==typeof e?e.options:e;if(t&&(u.render=t,u.staticRenderFns=n,u._compiled=!0),o&&(u.functional=!0),i&&(u._scopeId="data-v-"+i),r?(a=function(e){(e=e||this.$vnode&&this.$vnode.ssrContext||this.parent&&this.parent.$vnode&&this.parent.$vnode.ssrContext)||"undefined"==typeof __VUE_SSR_CONTEXT__||(e=__VUE_SSR_CONTEXT__),s&&s.call(this,e),e&&e._registeredComponents&&e._registeredComponents.add(r)},u._ssrRegister=a):s&&(a=c?function(){s.call(this,this.$root.$options.shadowRoot)}:s),a)if(u.functional){u._injectStyles=a;var l=u.render;u.render=function(e,t){return a.call(t),l(e,t)}}else{var p=u.beforeCreate;u.beforeCreate=p?[].concat(p,a):[a]}return{exports:e,options:u}}var w=m(_,h,[],!1,null,null,null);w.options.__file="src/pages/index.vue";var k=w.exports,g=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticClass:"new-spec-page"},[n("label",[e._v("\n\t\tName:\n\t\t"),n("el-input",{ref:"nameInput",attrs:{clearable:""},model:{value:e.newSpecName,callback:function(t){e.newSpecName=t},expression:"newSpecName"}})],1),e._v(" "),n("label",[e._v("\n\t\tDescription:\n\t\t"),n("el-input",{attrs:{type:"textarea",autosize:{minRows:2}},model:{value:e.newSpecDesc,callback:function(t){e.newSpecDesc=t},expression:"newSpecDesc"}})],1),e._v(" "),n("el-button",{on:{click:function(t){return e.cancel()}}},[e._v("Cancel")]),e._v(" "),n("el-button",{attrs:{disabled:e.disableCreate,type:"primary"},on:{click:function(t){return e.create()}}},[e._v("Create")])],1)};g._withStripped=!0;function S(e){return c.a.get("/ajax/spec",{specId:e}).fail(b)}var x={data:function(){return{newSpecName:"",newSpecDesc:""}},computed:{disableCreate:function(){return!this.newSpecName.trim()}},beforeRouteEnter:function(e,t,n){n((function(e){return e.nextTickFocusNameInput()}))},beforeRouteUpdate:function(e,t,n){this.newSpecName="",this.newSpecDesc="",this.nextTickFocusNameInput(),n()},methods:{nextTickFocusNameInput:function(){var e=this;this.$nextTick((function(){c()("input",e.$refs.nameInput.$el).focus()}))},cancel:function(){this.$router.push({name:"index"})},create:function(){var e,t,n=this;this.disableCreate||(e=this.newSpecName,t=this.newSpecDesc,c.a.post("/ajax/spec/create-spec",{name:e,desc:t}).fail(b)).then((function(e){n.$router.push({name:"spec",params:{specId:e}})}))}}},y=(n(14),m(x,g,[],!1,null,null,null));y.options.__file="src/pages/new-spec.vue";var $=y.exports,C=function(){var e=this,t=e.$createElement,n=e._self._c||t;return e.spec?n("div",{staticClass:"spec-page"},[n("h2",[e._v(e._s(e.spec.name))]),e._v(" "),e.spec.desc?n("div",{staticClass:"desc"},[e._v(e._s(e.spec.desc))]):e._e(),e._v(" "),n("p",[e._v("Owner: "+e._s(e.spec.ownerType)+" "+e._s(e.spec.ownerId))]),e._v(" "),n("p",[e._v("Created: "+e._s(e.spec.created))]),e._v(" "),e.spec.public?n("p",[e._v("Public")]):n("p",[e._v("Not public")]),e._v(" "),n("hr"),e._v(" "),n("spec-view",{attrs:{spec:e.spec}})],1):e._e()};C._withStripped=!0;var I=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticClass:"spec-view"},[n("ul",e._l(e.blocks,(function(t){return n("spec-block",{key:t.id,attrs:{block:t},on:{"prompt-add-subblock":e.promptAddSubblock}})})),1),e._v(" "),n("button",{on:{click:function(t){return e.promptAddBlock()}}},[e._v("Add block")]),e._v(" "),n("add-block-modal",{ref:"addBlockModal",attrs:{"spec-id":e.spec.id}})],1)};I._withStripped=!0;var j=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("li",{class:e.classes},[e.block.title?n("div",{staticClass:"title"},[e._v(e._s(e.block.title))]):e._e(),e._v(" "),e.block.body?n("div",{staticClass:"body"},[e._v(e._s(e.block.body))]):e._e(),e._v(" "),e.subblocks.length?n("ul",e._l(e.subblocks,(function(t){return n("spec-block",{key:t.id,attrs:{block:t},on:{"prompt-add-subblock":e.raisePromptAddSubblock}})})),1):e._e(),e._v(" "),n("button",{on:{click:function(t){return e.promptAddSubblock()}}},[e._v("Add subblock")])])};j._withStripped=!0;var N={name:"spec-block",props:{block:Object},data:function(){return{subblocks:this.block.subblocks?this.block.subblocks.slice():[]}},computed:{classes:function(){return["spec-block",this.block.type]}},methods:{promptAddSubblock:function(){var e=this;this.raisePromptAddSubblock(null,this.block.id,-1,(function(t){e.subblocks.push(t)}))},raisePromptAddSubblock:function(e,t,n,o){this.$emit("prompt-add-subblock",e,t,n,o)}}},A=(n(15),m(N,j,[],!1,null,null,null));A.options.__file="src/spec/block.vue";var T=A.exports,E=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("el-dialog",{staticClass:"spec-add-block-modal",attrs:{title:"Add block",visible:e.showing,width:e.$store.getters.dialogTinyWidth,"close-on-click-modal":!1},on:{"update:visible":function(t){e.showing=t},closed:function(t){return e.closed()}}},[n("label",[e._v("\n\t\tTitle\n\t\t"),n("el-input",{ref:"titleInput",attrs:{clearable:""},model:{value:e.title,callback:function(t){e.title=t},expression:"title"}})],1),e._v(" "),n("label",[e._v("\n\t\tBody\n\t\t"),n("el-input",{attrs:{type:"textarea",autosize:{minRows:2}},model:{value:e.body,callback:function(t){e.body=t},expression:"body"}})],1),e._v(" "),n("span",{directives:[{name:"loading",rawName:"v-loading",value:e.sending,expression:"sending"}],staticClass:"dialog-footer",attrs:{slot:"footer"},slot:"footer"},[n("el-button",{on:{click:function(t){e.showing=!1}}},[e._v("Cancel")]),e._v(" "),n("el-button",{attrs:{type:"primary",disabled:e.disableSubmit},on:{click:function(t){return e.submit()}}},[e._v("Add")])],1)])};E._withStripped=!0;var R={props:{specId:Number},data:function(){return{title:"",body:"",subspaceId:null,parentId:null,insertAt:null,callback:null,showing:!1,sending:!1}},computed:{disableSubmit:function(){return!this.title.trim()}},methods:{show:function(e,t,n,o){var s=this;this.subspaceId=e,this.parentId=t,this.insertAt=n,this.callback=o,this.showing=!0,this.$nextTick((function(){c()("input",s.$refs.titleInput.$el).focus()}))},submit:function(){var e=this;if(!this.disableSubmit){this.sending=!0;var t,n,o,s,i,r,a,u,l=this.callback;(t=this.specId,n=this.subspaceId,o=this.parentId,s=this.insertAt,i="text",r=null,a=this.title,u=this.body,c.a.post("/ajax/spec/create-block",{specId:t,subspaceId:n,parentId:o,insertAt:s||0===s?s:-1,refType:i,refId:r,title:a,body:u}).fail(b)).then((function(t){l(t),e.showing=!1,e.sending=!1})).fail((function(){e.sending=!1}))}},closed:function(){this.callback=null,this.title="",this.body=""}}},W=(n(16),m(R,E,[],!1,null,null,null));W.options.__file="src/spec/add-block-modal.vue";var O={components:{SpecBlock:T,AddBlockModal:W.exports},props:{spec:Object},data:function(){return{blocks:this.spec.blocks?this.spec.blocks.slice():[]}},methods:{promptAddBlock:function(){var e=this;this.$refs.addBlockModal.show(null,null,-1,(function(t){e.blocks.push(t)}))},promptAddSubblock:function(e,t,n,o){this.$refs.addBlockModal.show(e,t,n,o)}}},B=(n(17),m(O,I,[],!1,null,null,null));B.options.__file="src/spec/view.vue";var P={components:{SpecView:B.exports},data:function(){return{spec:null}},beforeRouteEnter:function(e,t,n){S(e.params.specId).then((function(e){n((function(t){t.spec=e}))})).fail((function(e){n({name:"ajax-error",params:{code:e.status},replace:!0})}))},beforeRouteUpdate:function(e,t,n){var o=this;S(e.params.specId).then((function(e){o.spec=e,n()})).fail((function(e){n({name:"ajax-error",params:{code:e.status},replace:!0})}))},methods:{}},D=(n(18),m(P,C,[],!1,null,null,null));D.options.__file="src/pages/spec.vue";var M=D.exports,V=function(){var e=this.$createElement;return(this._self._c||e)("div",[0===parseInt(this.$route.params.code,10)?[this._v("\n\t\tCould not connect to server\n\t")]:[this._v("\n\t\tRequest failed with error code "+this._s(this.$route.params.code)+"\n\t")]],2)};V._withStripped=!0;var U=m({},V,[],!1,null,null,null);U.options.__file="src/pages/ajax-error.vue";var z=U.exports,F=function(){var e=this.$createElement;return(this._self._c||e)("div",[this._v("\n\tNot found\n")])};F._withStripped=!0;var L=m({},F,[],!1,null,null,null);L.options.__file="src/pages/not-found.vue";var q=L.exports,X={store:d,router:new v.a({mode:"history",routes:[{name:"index",path:"/",component:k},{name:"new-spec",path:"/new-spec",component:$},{name:"spec",path:"/spec/:specId",component:M},{name:"ajax-error",path:"/ajax-error/:code",component:z},{path:"*",component:q}]}),computed:{username:function(){return this.$store.getters.username}},methods:{gotoIndex:function(){"index"!==this.$route.name&&this.$router.push({name:"index"})},logout:function(){window.location.href="/logout"}}},Y=(n(19),m(X,i,[],!1,null,null,null));Y.options.__file="src/app.vue";var Q=Y.exports;new s.a(Q).$mount("#app")}]);
<template>
<section v-if="subspec" class="subspec-page">

	<header>
		<div class="spec" @click="gotoSpec()">
			<h2>
				{{subspec.specName}}
			</h2>
		</div>
		<div class="subspec">
			<h3>
				<div class="right">
					<el-button @click="openManageSubspec()" size="mini" icon="el-icon-setting"/>
				</div>
				{{name}}
			</h3>
			<div v-if="desc" class="desc">{{desc}}</div>
		</div>
	</header>

	<spec-view :key="subspec.id" :subspec="subspec"/>

	<edit-subspec-modal ref="editSubspecModal" :spec-id="subspec.specId"/>

</section>
</template>

<script>
import $ from 'jquery';
import SpecView from '../spec/view.vue';
import EditSubspecModal from '../spec/edit-subspec-modal.vue';
import {ajaxLoadSubspec} from '../spec/ajax.js';
import {setWindowSubtitle} from '../utils.js';

export default {
	components: {
		SpecView,
		EditSubspecModal,
	},
	data() {
		return {
			subspec: null,
			name: '',
			desc: '',
		};
	},
	beforeRouteEnter(to, from, next) {
		ajaxLoadSubspec(to.params.specId, to.params.subspecId).then(subspec => {
			next(vm => {
				vm.setSubspec(subspec);
			});
		}).fail(jqXHR => {
			next({name: 'ajax-error', params: {code: jqXHR.status}, replace: true});
		});
	},
	beforeRouteUpdate(to, from, next) {
		// Reload spec even if same across navigation as view must be rebuilt using latest state
		ajaxLoadSubspec(to.params.specId, to.params.subspecId).then(subspec => {
			this.setSubspec(subspec);
			next();
			this.$nextTick(this.restoreScroll);
		}).fail(jqXHR => {
			next({name: 'ajax-error', params: {code: jqXHR.status}, replace: true});
		});
	},
	beforeRouteLeave(to, from, next) {
		setWindowSubtitle(); // clear
		next();
	},
	methods: {
		setSubspec(subspec) {
			this.subspec = subspec;
			this.name = subspec.name;
			this.desc = subspec.desc;
			setWindowSubtitle(subspec.name);
			// vue-router scrollBehavior is applied before spec-view has a chance to populate,
			// so restore the scroll position again after fully rendering.
			this.$nextTick(this.restoreScroll);
		},
		gotoSpec() {
			this.$router.push({name: 'spec', params: {specId: this.subspec.specId}});
		},
		openManageSubspec() {
			this.$refs.editSubspecModal.showEdit({
				id: this.subspec.id,
				name: this.name,
				desc: this.desc,
				created: this.subspec.created,
			}, updatedSubspec => {
				this.name = updatedSubspec.name;
				this.desc = updatedSubspec.desc;
				setWindowSubtitle(updatedSubspec.name);
			});
		},
		restoreScroll() {
			let position = this.$store.state.savedScrollPosition;
			if (position) {
				$(window).scrollTop(position.y).scrollLeft(position.x);
			}
		},
	},
};
</script>

<style lang="scss">
@import '../styles/_colours.scss';

.subspec-page {
	>header {
		margin-top: -1cm;
		margin-left: -1cm;
		margin-right: -1cm;
		margin-bottom: 1cm;
		>.spec {
			padding: 0.5cm 1cm;
			background-color: $spec;
			color: white;
			cursor: pointer;
			>h2 {
				margin: 0;
				padding: 0;
			}
		}
		>.subspec {
			padding: 0.5cm 1cm;
			background-color: $subspec;
			color: white;
			>h3 {
				margin: 0;
				padding: 0;
				>.right {
					float: right;
					font-size: small;
					margin-left: 20px;
					>*+* {
						margin-left: 10px;
					}
				}
			}
			>.desc {
				white-space: pre-wrap;
				margin-top: 10px;
			}
		}
	}
}
</style>
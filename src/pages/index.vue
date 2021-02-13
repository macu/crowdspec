<template>
<div class="index-page">

	<el-button
		@click="promptCreateSpec()"
		type="primary">
		New spec
	</el-button>

	<el-button
		@click="$router.push({name: 'community-review'})"
		type="default">
		Community review
	</el-button>

	<el-button
		v-if="$store.getters.userIsAdmin"
		@click="$router.push({name: 'admin'})"
		class="admin-access">
		Admin
	</el-button>

	<div class="user-specs">
		<h2>Your specs</h2>
		<p v-if="loading">Loading...</p>
		<ul v-else-if="userSpecs && userSpecs.length" class="specs-list">
			<router-link
				v-for="s in userSpecs"
				:key="s.id"
				tag="li"
				:to="{name: 'spec', params: {specId: s.id}}">
				<div class="info">
					<span class="status" :class="{public: s.public}">
						<template v-if="s.public"><i class="el-icon-umbrella"></i> Public</template>
						<template v-else><i class="el-icon-lock"></i> Private</template>
					</span>
				</div>
				<router-link :to="{name: 'spec', params: {specId: s.id}}" class="name">{{s.name}}</router-link>
				<div class="info">
					<span class="updated">Last modified <strong><moment :datetime="s.updated" :offset="true"/></strong></span>
				</div>
				<div v-if="s.desc" class="desc">{{s.desc}}</div>
			</router-link>
		</ul>
		<p v-else>You do not have any specs.</p>
	</div>

	<div class="public-specs" :class="{'adjust-margin': !!userSpecs.length}">
		<h2>Public specs</h2>
		<p v-if="loading">Loading...</p>
		<ul v-else-if="publicSpecs && publicSpecs.length" class="specs-list">
			<router-link
				v-for="s in publicSpecs"
				:key="s.id"
				tag="li"
				:to="{name: 'spec', params: {specId: s.id}}">
				<div class="info">
					<username :username="s.username" :highlight="s.highlight"/>
				</div>
				<router-link :to="{name: 'spec', params: {specId: s.id}}" class="name">{{s.name}}</router-link>
				<div class="info">
					<span class="updated">Last modified <strong><moment :datetime="s.updated" :offset="true"/></strong></span>
				</div>
				<!--<div class="tags">
					<el-tag size="mini" :closable="false" type="success">dsf</el-tag>
					<el-tag size="mini" :closable="false" type="primary">d</el-tag>
					<el-tag size="mini" :closable="false">asdfsadf</el-tag>
					<el-tag size="mini" :closable="false">asd</el-tag>
					<el-tag size="mini" :closable="false">fghfdghsdfg</el-tag>
					<el-tag size="mini" :closable="false">rtrgegrae</el-tag>
				</div>-->
				<div v-if="s.desc" class="desc">{{s.desc}}</div>
			</router-link>
		</ul>
		<p v-else>There are no public specs......</p>
	</div>

	<edit-spec-modal ref="editSpecModal"/>

</div>
</template>

<script>
import $ from 'jquery';
import Moment from '../widgets/moment.vue';
import Username from '../widgets/username.vue';
import EditSpecModal from '../spec/edit-spec-modal.vue';
import {alertError} from '../utils.js';

export default {
	components: {
		Moment,
		Username,
		EditSpecModal,
	},
	data() {
		return {
			userSpecs: [],
			publicSpecs: [],
			loading: true,
		};
	},
	mounted() {
		this.reloadSpecs();
	},
	beforeRouteUpdate(to, from, next) {
		this.reloadSpecs();
		next();
	},
	methods: {
		reloadSpecs() {
			this.loading = true;
			$.get('/ajax/home').then(payload => {
				this.userSpecs = payload.userSpecs;
				this.publicSpecs = payload.publicSpecs;
				this.loading = false;
			}).fail(jqXHR => {
				this.loading = false;
				alertError(jqXHR);
			});
		},
		promptCreateSpec() {
			this.$refs.editSpecModal.showCreate(newSpecId => {
				this.$router.push({name: 'spec', params: {specId: newSpecId}});
			});
		},
	},
};
</script>

<style lang="scss">
@import '../_styles/_breakpoints.scss';
@import '../_styles/_colours.scss';
@import '../_styles/_app.scss';

.index-page {

	padding: $content-area-padding;

	@include mobile {
		padding: $content-area-padding-sm;
	}

	.user-specs, .public-specs {
		margin-top: $content-area-padding;
		@include mobile {
			margin-top: $content-area-padding-sm;
		}
	}

	ul.specs-list {

		>li {
			padding: 20px;
			cursor: pointer;

			&:not(:first-child) {
				margin-top: 1px;
			}

			&:hover {
				background-color: $shadow-bg;
			}

			>.info {
				font-size: small;

				>span {
					display: inline-block;
					margin-right: 20px;

					@include mobile {
						display: block;
					}

					&.status {
						color: gray;

						&.public {
							color: green;
							font-weight: green;
						}

						>i {
							display: inline-block;
							font-weight: bold;
							margin-right: $icon-spacing;
						}
					}
				}
			}

			>.name {
				display: block;
				font-size: larger;
			}

			>.desc {
				font-size: .7rem;
				line-height: 1rem;
				white-space: pre-wrap;

				// Text hidden beyond 3 lines
				max-height: 3rem;
				overflow: hidden;

				// Special behaviour for Chrome
				display: -webkit-box;
				-webkit-line-clamp: 3;
				-webkit-box-orient: vertical;
				text-overflow: ellipsis;
			}

			>* + * {
				margin-top: 10px;
			}

		}
	} // ul.specs-list

	.el-button.admin-access {
		color: white;
		background-color: $admin-bg;
		border-color: $admin-bg;
	}

} // .index-page
</style>

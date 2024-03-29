<template>
<div class="index-page">

	<div v-if="loggedIn" class="actions">
		<el-button
			@click="promptCreateSpec()"
			type="primary">
			New spec
		</el-button>

		<el-button
			@click="$router.push({name: 'community-review'})"
			:type="unreadCommunity ? 'primary' : 'default'">
			Community review
			<template v-if="unreadCommunity">({{unreadCommunity}} unread)</template>
		</el-button>

		<el-button
			v-if="$store.getters.userIsAdmin"
			@click="$router.push({name: 'admin'})"
			class="admin-access">
			Admin
		</el-button>
	</div>

	<div v-if="loggedIn" class="user-specs">
		<h2>Your specs</h2>
		<p v-if="loading">
			<loading-message message="Loading..."/>
		</p>
		<ul v-else-if="userSpecs && userSpecs.length"
			class="specs-list"
			:class="specsLayoutList ? 'list' : 'grid'">
			<router-link
				v-for="s in userSpecs"
				:key="s.id"
				:to="{name: 'spec', params: {specId: s.id}}"
				custom
				v-slot="{ navigate, href }">
				<li @click="navigate">
					<div>
						<div class="info">
							<span class="status" :class="{public: s.public}">
								<template v-if="s.public"><i class="material-icons">beach_access</i> Public</template>
								<template v-else><i class="material-icons">lock</i> Private</template>
							</span>
						</div>
						<a :href="href" @click="navigate" class="name">{{s.name}}</a>
						<div class="info">
							<span class="updated">Last modified <strong><moment :datetime="s.updated" :offset="true"/></strong></span>
						</div>
						<div v-if="s.desc" class="desc">{{s.desc}}</div>
					</div>
				</li>
			</router-link>
		</ul>
		<p v-else>You do not have any specs.</p>
	</div>

	<div class="public-specs" :class="{'adjust-margin': !!userSpecs.length}">
		<h2>Public specs</h2>
		<p v-if="loading">
			<loading-message message="Loading..."/>
		</p>
		<ul v-else-if="publicSpecs && publicSpecs.length"
			class="specs-list"
			:class="specsLayoutList ? 'list' : 'grid'">
			<router-link
				v-for="s in publicSpecs"
				:key="s.id"
				:to="{name: 'spec', params: {specId: s.id}}"
				custom
				v-slot="{ navigate, href }">
				<li @click="navigate">
					<div>
						<div class="info">
							<username :username="s.username" :highlight="s.highlight"/>
						</div>
						<a :href="href" @click="navigate" class="name">{{s.name}}</a>
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
					</div>
				</li>
			</router-link>
		</ul>
		<p v-else>There are no public specs......</p>
	</div>

	<edit-spec-modal v-if="loggedIn" ref="editSpecModal"/>

</div>
</template>

<script>
import $ from 'jquery';
import Moment from '../widgets/moment.vue';
import Username from '../widgets/username.vue';
import EditSpecModal from '../spec/edit-spec-modal.vue';
import LoadingMessage from '../widgets/loading.vue';
import {alertError} from '../utils.js';

export default {
	components: {
		Moment,
		Username,
		EditSpecModal,
		LoadingMessage,
	},
	data() {
		return {
			unreadCommunity: 0,
			userSpecs: [],
			publicSpecs: [],
			loading: true,
		};
	},
	computed: {
		loggedIn() {
			return this.$store.getters.loggedIn;
		},
		specsLayoutList() {
			return this.$store.getters.specsLayoutList;
		},
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
				this.unreadCommunity = payload.unread;
				this.userSpecs = payload.userSpecs;
				this.publicSpecs = payload.publicSpecs;
				this.loading = false;
			}).fail(jqXHR => {
				this.loading = false;
				alertError(jqXHR);
			});
		},
		promptCreateSpec() {
			if (!this.$store.getters.loggedIn) {
				return;
			}
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

	.actions {
		margin-top: -10px;
		.el-button {
			margin-top: 10px;
		}
		.el-button.admin-access {
			@include custom-button(white, $admin-bg);
		}
	}

	*+.user-specs, *+.public-specs {
		margin-top: 70px;
		@include mobile {
			margin-top: 60px;
		}
	}

	ul.specs-list {

		&.list {
			>li {
				&:not(:first-child) {
					margin-top: 1px;
				}
			}
		}

		&.grid {
			padding: 0;
			display: grid;
			grid-template-columns: 1fr;
			align-items: start;
			column-gap: 20px;
			row-gap: 20px;

			>li {
				display: inline-block;
				background-color: scale-color($shadow-bg, $lightness: 60%);
			}

			@media (min-width: $min-lg) {
				grid-template-columns: 1fr 1fr;
			}

			@media (min-width: $min-xl) {
				grid-template-columns: 1fr 1fr 1fr;
			}
		}

		>li {
			padding: 20px;
			cursor: pointer;

			&:hover {
				background-color: $shadow-bg;
			}

			>div {

				>.name {
					display: block;
					font-size: larger;
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

		}
	} // ul.specs-list

} // .index-page
</style>

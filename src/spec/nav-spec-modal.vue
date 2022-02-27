<template>
<el-dialog
	title="Navigate spec"
	v-model="showing"
	:width="$store.getters.dialogTinyWidth"
	@closed="closed()"
	custom-class="nav-spec-modal">

	<p v-if="loading">
		<loading-message message="Loading..."/>
	</p>

	<template v-else-if="subspecs.length">

		<div v-if="subspecs.length > 10" class="filter">
			<el-input
				type="text"
				v-model="filter"
				placeholder="Filter by name or description"
				clearable
				/>
		</div>

		<template v-if="filteredSubspecs.length">
			<ref-subspec v-for="s in subspecs" :key="s.id" :item="s"/>
		</template>
		<p v-else><em>No matching subspecs.</em></p>

	</template>
	<p v-else><em>No subspecs.</em></p>

	<el-button
		v-if="enableEditing"
		@click="openCreateSubspec()"
		class="new-subspec-button">
		New subspec
	</el-button>

	<template #footer>
		<span class="dialog-footer">
			<el-button @click="showing = false">Close</el-button>
			<el-button
				v-if="subspecId"
				@click="goToSpec()"
				type="primary">
				Go to spec
			</el-button>
		</span>
	</template>

</el-dialog>
</template>

<script>
import RefSubspec from './ref-subspec.vue';
import LoadingMessage from '../widgets/loading.vue';
import {alertError} from '../utils.js';

export default {
	components: {
		RefSubspec,
		LoadingMessage,
	},
	props: {
		specId: Number,
		subspecId: Number,
		enableEditing: Boolean,
	},
	emits: ['open-create-subspec'],
	data() {
		return {
			subspecs: [],
			loading: false,
			showing: false,
			filter: '',
		};
	},
	computed: {
		filteredSubspecs() {
			if (!this.filter) {
				return this.subspecs;
			}
			let filter = this.filter.toLowerCase();
			let filtered = [];
			for (let i = 0; i < this.subspecs.length; i++) {
				let subspec = this.subspecs[i];
				if (
					subspec.name.toLowerCase().indexOf(filter) >= 0 ||
					(subspec.desc && subspec.desc.toLowerCase().indexOf(filter) >= 0)
				) {
					filtered.push(subspec);
				}
			}
			return filtered;
		},
	},
	watch: {
		'$route': {
			deep: true,
			handler() {
				// Hide modal on route changes
				this.showing = false;
			},
		},
	},
	methods: {
		show() {
			this.loading = true;
			this.showing = true;
			$.get('/ajax/spec/subspecs', {
				specId: this.specId,
			}).then(subspecs => {
				this.subspecs = subspecs;
				this.loading = false;
			}).fail(error => {
				this.loading = false;
				this.showing = false;
				alertError(error);
			})
		},
		openCreateSubspec() {
			this.$emit('open-create-subspec');
			this.showing = false;
		},
		goToSpec() {
			if (
				this.$route.name !== 'spec'
			) {
				this.$router.push({
					name: 'spec',
					params: {
						specId: this.specId,
					},
				});
			}
			this.showing = false;
		},
		closed() {
			this.loading = false;
			this.subspecs = [];
			this.filter = '';
		},
	},
};
</script>

<style lang="scss">
.nav-spec-modal {
	.el-select {
		width: 100%;
	}
	.ref-subspec+.ref-subspec {
		margin-top: 10px;
	}
	.new-subspec-button {
		width: 100%;
		margin-top: 20px;
	}
	.filter {
		display: block;
		margin-bottom: 20px;
		// >div:first-child:not([class]) {
		// 	font-weight: bold;
		// 	margin-bottom: 5px;
		// }
		>.el-input {
			width: 100%;
		}
	}
	p>em {
		color: gray;
	}
}
</style>

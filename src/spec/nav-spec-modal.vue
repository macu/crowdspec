<template>
<el-dialog
	title="Navigate spec"
	:visible.sync="showing"
	:width="$store.getters.dialogTinyWidth"
	@closed="closed()"
	class="nav-spec-modal">

	<p v-if="loading">Loading...</p>

	<template v-else-if="subspecs.length">

		<ref-subspec v-for="s in subspecs" :key="s.id" :item="s"/>

	</template>

	<p v-else>No subspecs.</p>

	<span slot="footer" class="dialog-footer">
		<el-button @click="showing = false">Close</el-button>
		<el-button
			v-if="subspecId"
			@click="goToSpec()"
			type="primary">
			Go to spec
		</el-button>
	</span>

</el-dialog>
</template>

<script>
import RefSubspec from './ref-subspec.vue';
import {alertError} from '../utils.js';

export default {
	components: {
		RefSubspec,
	},
	props: {
		specId: Number,
		subspecId: Number,
	},
	data() {
		return {
			subspecs: [],
			loading: false,
			showing: false,
		};
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
}
</style>

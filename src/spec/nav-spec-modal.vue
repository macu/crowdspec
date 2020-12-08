<template>
<el-dialog
	title="Navigate spec"
	:visible.sync="showing"
	:width="$store.getters.dialogTinyWidth"
	@closed="closed()"
	class="nav-spec-modal">

	<p v-if="loading">Loading...</p>

	<el-select v-else-if="subspecs.length" v-model="selectedSubspecId" placeholder="Choose subspec">
		<el-option v-for="s in subspecs" :key="s.id" :value="s.id" :label="s.name"/>
	</el-select>

	<p v-else>No subspecs.</p>

	<span slot="footer" class="dialog-footer">
		<el-button @click="showing = false">Close</el-button>
		<el-button
			v-if="subspecId"
			@click="goToSpec()"
			type="primary">
			Go to spec
		</el-button>
		<el-button
			v-if="subspecs.length"
			@click="goToSubspec()"
			type="primary"
			:disabled="disableGoToSubspec">
			Go to subspec
		</el-button>
	</span>

</el-dialog>
</template>

<script>
import {alertError} from '../utils.js';

export default {
	props: {
		specId: Number,
		subspecId: Number,
	},
	data() {
		return {
			subspecs: [],
			selectedSubspecId: null,
			loading: false,
			showing: false,
		};
	},
	computed: {
		disableGoToSubspec() {
			return this.loading || !this.subspecs.length || !this.selectedSubspecId;
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
			this.$router.push({
				name: 'spec',
				params: {
					specId: this.specId,
				},
			});
			this.showing = false;
		},
		goToSubspec() {
			this.$router.push({
				name: 'subspec',
				params: {
					specId: this.specId,
					subspecId: this.selectedSubspecId,
				},
			});
			this.showing = false;
		},
		closed() {
			this.loading = false;
			this.subspecs = [];
			this.selectedSubspecId = null;
		},
	},
};
</script>

<style lang="scss">
.nav-spec-modal {
	.el-select {
		width: 100%;
	}
}
</style>

<template>
<div class="ref-subspec-form">
	<el-card v-if="creating">
		<label>
			New subspec name
			<el-input ref="newSubspecNameInput" v-model="subspecName" :maxlength="nameMaxLength" clearable/>
		</label>
		<label>
			Description
			<el-input type="textarea" v-model="subspecDesc" :autosize="{minRows: 2}"/>
		</label>
		<el-button v-if="selectModeAvailable" @click="creating=false" size="small">
			Cancel
		</el-button>
	</el-card>
	<template v-else-if="selecting">
		<p v-if="loading">Loading subspecs...</p>
		<template v-else-if="subspecs && subspecs.length">
			<el-select v-model="refId" filterable placeholder="Select subspec">
				<el-option v-for="o in subspecs" :key="o.id" :value="o.id" :label="o.name"/>
			</el-select>
			<ref-subspec v-if="selectedSubspec" :item="selectedSubspec"/>
		</template>
		<el-button @click="creating=true" size="small">
			Create new subspec
		</el-button>
	</template>
	<template v-else>
		<ref-subspec v-if="initialSubspec" :item="initialSubspec"/>
		<div>
			<el-button @click="creating=true" size="small">
				Create new subspec
			</el-button>
			<el-button v-if="selectModeAvailable" @click="selecting=true" size="small">
				Select a different subspec
			</el-button>
		</div>
	</template>
</div>
</template>

<script>
import RefSubspec from './ref-subspec.vue';
import {ajaxLoadSubspecs} from './ajax.js';

export default {
	components: {
		RefSubspec,
	},
	props: {
		specId: Number,
		initialSubspec: Object,
		valid: Boolean, // sync
		fields: Object, // sync
	},
	data() {
		return {
			// user inputs
			subspecName: '',
			subspecDesc: '',
			refId: this.initialSubspec ? this.initialSubspec.id : null,
			// state
			creating: false,
			selecting: !this.initialSubspec,
			subspecs: null, // null indicates no fetch done yet
			loading: false,
		};
	},
	computed: {
		selectModeAvailable() {
			// Allow cancelling if haven't yet loaded spec links or there are some
			return this.subspecs === null || this.subspecs.length;
		},
		selectedSubspec() {
			if (this.subspecs && this.refId) {
				for (var i = 0; i < this.subspecs.length; i++) {
					if (this.subspecs[i].id === this.refId) {
						return this.subspecs[i];
					}
				}
			}
			return null;
		},
		isValid() {
			return this.creating ? !!this.subspecName.trim() : !!this.refId;
		},
		refFields() {
			return this.creating ? {
				refName: this.subspecName,
				refDesc: this.subspecDesc,
			} : {refId: this.refId};
		},
		nameMaxLength() {
			return window.const.specNameMaxLength;
		},
	},
	watch: {
		selecting: {
			immediate: true,
			handler(selecting) {
				if (selecting) {
					this.loadSubspecs();
				}
			},
		},
		isValid: {
			immediate: true,
			handler(valid) {
				// update sync prop value
				this.$emit('update:valid', valid);
			},
		},
		refFields: {
			immediate: true,
			handler(fields) {
				// update sync prop value
				this.$emit('update:fields', fields);
			},
		},
	},
	methods: {
		loadSubspecs() {
			if (this.loading) {
				return;
			}
			this.loading = true;
			ajaxLoadSubspecs(this.specId).then(subspecs => {
				this.subspecs = subspecs;
				if (!subspecs.length) {
					this.creating = true;
				}
				this.loading = false;
			}).fail(() => {
				this.subspecs = [];
				this.creating = true;
				this.loading = false;
			});
		},
	},
};
</script>

<style lang="scss">
.ref-subspec-form {
	>*+* {
		margin-top: 10px;
	}
	>.el-card {
		>.el-card__body {
			>*+* {
				margin-top: 10px;
			}
			>label {
				display: block;
				>.el-input {
					display: block;
					width: 100%;
				}
			}
		}
	}
	>.el-select {
		display: block;
		width: 100%;
	}
}
</style>

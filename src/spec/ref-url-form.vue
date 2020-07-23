<template>
<div class="ref-url-form">
	<el-card v-if="creating">
		<label>
			New link URL
			<el-input v-model="newUrl" clearable/>
		</label>
		<el-button v-if="selectModeAvailable" @click="creating=false" size="small">
			Cancel
		</el-button>
	</el-card>
	<template v-else-if="selecting">
		<p v-if="loading">Loading links...</p>
		<template v-else-if="urlObjects && urlObjects.length">
			<el-select v-model="refId" filterable placeholder="Select link">
				<el-option v-for="o in urlObjects" :key="o.id" :value="o.id" :label="o.title || o.url"/>
			</el-select>
			<ref-url
				v-if="selectedUrlObject"
				:item="selectedUrlObject"
				show-edit
				@edit="openEditUrl(selectedUrlObject)"
				/>
		</template>
		<el-button @click="creating=true" size="small">
			Create new link
		</el-button>
	</template>
	<template v-else>
		<ref-url
			v-if="initialUrlObject"
			:item="initialUrlObject"
			show-edit
			@edit="openEditUrl(initialUrlObject)"
			/>
		<div>
			<el-button @click="creating=true" size="small">
				Create new link
			</el-button>
			<el-button v-if="selectModeAvailable" @click="selecting=true" size="small">
				Select a different link
			</el-button>
		</div>
	</template>
</div>
</template>

<script>
import RefUrl from './ref-url.vue';
import {ajaxLoadUrls} from './ajax.js';
import {isValidURL} from '../utils.js';

export default {
	components: {
		RefUrl,
	},
	props: {
		specId: Number,
		initialUrlObject: Object,
		valid: Boolean, // sync
		fields: Object, // sync
	},
	data() {
		return {
			// user inputs
			newUrl: '',
			refId: this.initialUrlObject ? this.initialUrlObject.id : null,
			// state
			creating: false,
			selecting: !this.initialUrlObject, // don't start in select mode if initial
			urlObjects: null, // null indicates not yet loaded
			loading: false,
		};
	},
	computed: {
		selectModeAvailable() {
			// Allow cancelling if haven't yet loaded spec links or there are some
			return this.urlObjects === null || this.urlObjects.length;
		},
		selectedUrlObject() {
			if (this.urlObjects && this.refId) {
				for (var i = 0; i < this.urlObjects.length; i++) {
					if (this.urlObjects[i].id === this.refId) {
						return this.urlObjects[i];
					}
				}
			}
			return null;
		},
		isValid() {
			return this.creating ? isValidURL(this.newUrl) : !!this.refId;
		},
		refFields() {
			return this.creating ? {refUrl: this.newUrl} : {refId: this.refId};
		},
	},
	watch: {
		selecting: {
			immediate: true,
			handler(selecting) {
				if (selecting) {
					this.loadLinks();
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
		loadLinks() {
			if (this.loading) {
				return;
			}
			this.loading = true;
			ajaxLoadUrls(this.specId).then(urls => {
				this.urlObjects = urls;
				if (!urls.length) {
					this.selecting = false;
					if (!this.initialUrlObject) {
						this.creating = true;
					}
				}
				this.loading = false;
			}).fail(() => {
				this.urlObjects = [];
				this.selecting = false;
				if (!this.initialUrlObject) {
					this.creating = true;
				}
				this.loading = false;
			});
		},
		openEditUrl(urlObject) {
			this.$emit('open-edit-url', urlObject, updatedUrlObject => {
				// Updated
				if (this.urlObjects) {
					for (var i = 0; i < this.urlObjects.length; i++) {
						if (this.urlObjects[i].id === urlObject.id) {
							this.urlObjects.splice(i, 1, updatedUrlObject); // Replace
							break;
						}
					}
				}
			}, () => {
				// Deleted
				if (this.urlObjects) {
					for (var i = 0; i < this.urlObjects.length; i++) {
						if (this.urlObjects[i].id === urlObject.id) {
							this.urlObjects.splice(i, 1); // Remove
							break;
						}
					}
				}
				if (this.refId === urlObject.id) {
					this.refId = null;
				}
			});
		},
	},
};
</script>

<style lang="scss">
.ref-url-form {
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

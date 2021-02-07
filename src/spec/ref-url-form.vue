<template>
<div class="ref-url-form">
	<el-card v-if="creating">
		<label>
			New link URL
			<el-input v-model="newUrl" :maxlength="urlMaxLength" clearable/>
			<el-alert v-if="newUrl && !isValid" title="Invalid URL" type="error" :closable="false"/>
			<el-alert v-else-if="newUrlIsVideo" title="Embeddable video detected" type="info" :closable="false"/>
		</label>
		<el-button v-if="selectModeAvailable" @click="() => { creating = false; selecting = true; }" size="small">
			Select an existing link
		</el-button>
		<el-button v-if="initialUrlObject" @click="() => { creating = false; selecting = false; }" size="small">
			Cancel
		</el-button>
	</el-card>
	<template v-else-if="selecting">
		<p v-if="loading">Loading links...</p>
		<template v-else-if="urlObjects && urlObjects.length">
			<el-select v-model="refId" filterable :filter-method="filterUrls" placeholder="Select link">
				<el-option v-for="o in filteredUrlObjects" :key="o.id" :value="o.id" :label="o.title || o.url"/>
			</el-select>
			<ref-url
				v-if="selectedUrlObject"
				:item="selectedUrlObject"
				show-edit
				@edit="openEditUrl(selectedUrlObject)"
				@play="raisePlayVideo(selectedUrlObject)"
				/>
		</template>
		<el-button @click="creating = true" size="small">
			Create new link
		</el-button>
		<el-button v-if="initialUrlObject" @click="selecting = false" size="small">
			Cancel
		</el-button>
	</template>
	<template v-else>
		<ref-url
			v-if="initialUrlObject"
			:item="initialUrlObject"
			show-edit
			@edit="openEditUrl(initialUrlObject)"
			@play="raisePlayVideo(initialUrlObject)"
			/>
		<div>
			<el-button @click="creating = true" size="small">
				Create new link
			</el-button>
			<el-button v-if="selectModeAvailable" @click="selecting = true" size="small">
				Select a different link
			</el-button>
		</div>
	</template>
</div>
</template>

<script>
import RefUrl from './ref-url.vue';
import {ajaxLoadUrls} from './ajax.js';
import {isValidURL, isVideoURL, debounce} from '../utils.js';

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
			filter: '',
			// state
			creating: !this.initialUrlObject, // start in create mode if no initial
			selecting: false,
			urlObjects: null, // null indicates not yet loaded
			loading: false,
		};
	},
	computed: {
		filteredUrlObjects() {
			if (this.filter && this.urlObjects) {
				let filter = this.filter.toLowerCase();
				let filtered = [];
				for (var i = 0; i < this.urlObjects.length; i++) {
					let o = this.urlObjects[i];
					if (o.url.toLowerCase().indexOf(filter) >= 0 ||
						(o.title && o.title.toLowerCase().indexOf(filter) >= 0)) {
						filtered.push(o);
					}
				}
				return filtered;
			}
			return this.urlObjects;
		},
		selectModeAvailable() {
			// Allow switching to select mode if spec links haven't been loaded
			// or there are options other than the initial urlObject
			return this.urlObjects === null || this.urlObjects.length > 1 ||
				(this.urlObjects.length && !this.initialUrlObject) ||
				(this.urlObjects.length === 1 &&
					this.urlObjects[0].id !== this.initialUrlObject.id);
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
		newUrlIsVideo() {
			return isValidURL(this.newUrl) && isVideoURL(this.newUrl);
		},
		isValid() {
			return this.creating ? isValidURL(this.newUrl) : !!this.refId;
		},
		refFields() {
			// refType is passed separately to ajaxCreateBlock/ajaxSaveBlock
			if (this.creating) {
				return {refUrl: this.newUrl};
			} else if (this.selecting) {
				return {refId: this.refId};
			} else if (this.initialUrlObject) {
				return {refId: this.initialUrlObject.id};
			}
			return null;
		},
		urlMaxLength() {
			return window.const.urlMaxLength;
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
			if (this.urlObjects || this.loading) {
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
		filterUrls(filter) {
			if (!filter) {
				this.filter = '';
			} else {
				if (!this.debouncedUpdateFilter) {
					this.debouncedUpdateFilter = debounce(filter => {
						this.filter = filter;
					});
				}
				this.debouncedUpdateFilter(filter);
			}
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
		raisePlayVideo(urlObject) {
			this.$emit('play-video', urlObject);
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
				>.el-alert {
					margin-top: 10px;
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

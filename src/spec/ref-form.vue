<template>
<div class="ref-form">

	<template v-if="refType === REF_TYPE_URL">
		<el-card v-if="urlMode === MODE_CREATE">
			<p v-if="creatingUrlObject">
				<loading-message message="Creating..."/>
			</p>
			<label v-else>
				<div>New link URL</div>
				<el-input
					v-model="newUrl"
					:maxlength="urlMaxLength"
					clearable
					/>
				<el-alert
					v-if="newUrl && !isValid"
					title="Invalid URL"
					type="error"
					:closable="false"
					/>
				<el-alert
					v-else-if="newUrlIsVideo"
					title="Embeddable video detected"
					type="info"
					:closable="false"
					/>
			</label>
			<el-button
				@click="createUrl()"
				:disabled="disableCreateUrl"
				type="primary">
				Create
			</el-button>
			<el-button v-if="urlSelectModeAvailable" @click="urlMode = MODE_SELECT">
				Select an existing link
			</el-button>
			<el-button v-if="initialUrlObject" @click="urlMode = MODE_KEEP">
				Cancel
			</el-button>
		</el-card>
		<template v-else-if="urlMode === MODE_SELECT">
			<p v-if="loadingUrlObjects">
				<loading-message message="Loading links..."/>
			</p>
			<template v-else-if="urlObjects && urlObjects.length">
				<el-select
					v-model="urlId"
					filterable
					:filter-method="filterUrlObjects"
					placeholder="Select link">
					<el-option
						v-for="o in filteredUrlObjects"
						:key="o.id"
						:value="o.id"
						:label="o.title || o.url"
						/>
				</el-select>
				<ref-url
					v-if="selectedUrlObject"
					:item="selectedUrlObject"
					show-edit
					@edit="openEditUrl(selectedUrlObject)"
					@play="raisePlayVideo(selectedUrlObject)"
					/>
			</template>
			<div key="url-select-actions">
				<el-button @click="urlMode = MODE_CREATE">
					Create new link
				</el-button>
				<el-button v-if="initialUrlObject" @click="urlMode = MODE_KEEP">
					Cancel
				</el-button>
			</div>
		</template>
		<template v-else>
			<ref-url
				v-if="initialUrlObject"
				:item="initialUrlObject"
				show-edit
				@edit="openEditUrl(initialUrlObject)"
				@play="raisePlayVideo(initialUrlObject)"
				/>
			<div key="url-keep-actions">
				<el-button @click="urlMode = MODE_CREATE">
					Create new link
				</el-button>
				<el-button v-if="urlSelectModeAvailable" @click="urlMode = MODE_SELECT">
					Select a different link
				</el-button>
			</div>
		</template>
	</template>

	<template v-else-if="refType === REF_TYPE_SUBSPEC">
		<el-card v-if="subspecMode === MODE_CREATE">
			<p v-if="creatingSubspec">
				<loading-message message="Creating..."/>
			</p>
			<template v-else>
				<label>
					<div>New subspec name</div>
					<el-input ref="newSubspecNameInput"
						v-model="newSubspecName"
						:maxlength="subspecNameMaxLength"
						clearable
						/>
				</label>
				<label>
					<div>Description</div>
					<el-input type="textarea"
						v-model="newSubspecDesc"
						:autosize="{minRows: 2}"
						/>
				</label>
			</template>
			<el-button
				@click="createSubspec()"
				:disabled="disableCreateSubspec"
				type="primary">
				Create
			</el-button>
			<el-button v-if="subspecSelectModeAvailable" @click="subspecMode = MODE_SELECT">
				Select an existing subspec
			</el-button>
			<el-button v-if="initialSubspec" @click="subspecMode = MODE_KEEP">
				Cancel
			</el-button>
		</el-card>
		<template v-else-if="subspecMode === MODE_SELECT">
			<p v-if="loadingSubspecs">
				<loading-message message="Loading subspecs..."/>
			</p>
			<template v-else-if="subspecs && subspecs.length">
				<el-select v-model="subspecId" filterable placeholder="Select subspec">
					<el-option v-for="o in subspecs" :key="o.id" :value="o.id" :label="o.name"/>
				</el-select>
				<ref-subspec v-if="selectedSubspec" :item="selectedSubspec"/>
			</template>
			<div key="subspec-select-actions">
				<el-button @click="subspecMode = MODE_CREATE" size="small">
					Create new subspec
				</el-button>
				<el-button v-if="initialSubspec" @click="subspecMode = MODE_KEEP" size="small">
					Cancel
				</el-button>
			</div>
		</template>
		<template v-else>
			<ref-subspec v-if="initialSubspec" :item="initialSubspec"/>
			<div key="subspec-keep-actions">
				<el-button @click="subspecMode = MODE_CREATE" size="small">
					Create new subspec
				</el-button>
				<el-button v-if="subspecSelectModeAvailable" @click="subspecMode = MODE_SELECT" size="small">
					Select a different subspec
				</el-button>
			</div>
		</template>
	</template>

</div>
</template>

<script>
import RefUrl from './ref-url.vue';
import RefSubspec from './ref-subspec.vue';
import LoadingMessage from '../widgets/loading.vue';
import {
	ajaxLoadUrls,
	ajaxLoadSubspecs,
	ajaxCreateUrl,
	ajaxCreateSubspec,
} from './ajax.js';
import {
	REF_TYPE_URL, REF_TYPE_SUBSPEC,
} from './const.js';
import {
	isValidURL, isVideoURL,
	debounce,
	notifySuccess, notifyInfo,
} from '../utils.js';

const MODE_KEEP = 'keep';
const MODE_CREATE = 'create';
const MODE_SELECT = 'select';

// The various ref forms are all managed by this one component
// so that inputs and selections are maintained if the user
// switches among ref types.

export default {
	components: {
		RefUrl,
		RefSubspec,
		LoadingMessage,
	},
	props: {
		specId: {
			type: Number,
			required: true,
		},
		refType: String, // selected in edit-block-modal
		existingRefType: String, // existing
		existingRefItem: Object, // existing
		valid: Boolean, // v-model:valid
		fields: Object, // v-model:fields
	},
	emits: ['update:valid', 'update:fields', 'open-edit-url', 'play-video'],
	data() {
		return {

			// URL
			newUrl: '',
			urlId: this.existingRefType === REF_TYPE_URL &&
				this.existingRefItem && this.existingRefItem.id || null,
			urlFilter: '',
			urlMode: (this.existingRefType === REF_TYPE_URL &&
				this.existingRefItem) ? MODE_KEEP : MODE_CREATE,
			urlObjects: null,
			urlObjectsLoaded: false,
			loadingUrlObjects: false,
			creatingUrlObject: false,

			// Subspec
			newSubspecName: '',
			newSubspecDesc: '',
			subspecId: this.existingRefType === REF_TYPE_SUBSPEC &&
				this.existingRefItem && this.existingRefItem.id || null,
			subspecFilter: '',
			subspecMode: (this.existingRefType === REF_TYPE_SUBSPEC &&
				this.existingRefItem) ? MODE_KEEP : MODE_CREATE,
			subspecs: null,
			subspecsLoaded: false,
			loadingSubspecs: false,
			creatingSubspec: false,

		};
	},
	computed: {

		// Constants
		REF_TYPE_URL() {
			return REF_TYPE_URL;
		},
		REF_TYPE_SUBSPEC() {
			return REF_TYPE_SUBSPEC;
		},
		MODE_KEEP() {
			return MODE_KEEP;
		},
		MODE_CREATE() {
			return MODE_CREATE;
		},
		MODE_SELECT() {
			return MODE_SELECT;
		},

		// URLs
		initialUrlObject() {
			return this.existingRefType === REF_TYPE_URL && this.existingRefItem;
		},
		urlSelectModeAvailable() {
			// Allow switching to select mode if spec links haven't been loaded
			// or there are options other than the initial urlObject
			return this.urlObjects === null || this.urlObjects.length > 1 ||
				(this.urlObjects.length && !this.initialUrlObject) ||
				(this.urlObjects.length === 1 &&
					this.urlObjects[0].id !== this.initialUrlObject.id);
		},
		filteredUrlObjects() {
			if (this.urlFilter && this.urlObjects) {
				let urlFilter = this.urlFilter.toLowerCase();
				let filtered = [];
				for (var i = 0; i < this.urlObjects.length; i++) {
					let o = this.urlObjects[i];
					if (o.url.toLowerCase().indexOf(urlFilter) >= 0 ||
						(o.title && o.title.toLowerCase().indexOf(urlFilter) >= 0)) {
						filtered.push(o);
					}
				}
				return filtered;
			}
			return this.urlObjects;
		},
		selectedUrlObject() {
			if (this.urlObjects && this.urlId) {
				for (var i = 0; i < this.urlObjects.length; i++) {
					if (this.urlObjects[i].id === this.urlId) {
						return this.urlObjects[i];
					}
				}
			}
			return null;
		},
		newUrlIsVideo() {
			return isValidURL(this.newUrl) && isVideoURL(this.newUrl);
		},
		disableCreateUrl() {
			return !isValidURL(this.newUrl) || this.creatingUrlObject;
		},
		urlMaxLength() {
			return window.const.urlMaxLength;
		},

		// Subspecs
		initialSubspec() {
			return this.existingRefType === REF_TYPE_SUBSPEC && this.existingRefItem;
		},
		subspecSelectModeAvailable() {
			// Allow switching to select mode if subspecs haven't been loaded
			// or there are options other than the intial subspec
			return this.subspecs === null || this.subspecs.length > 1 ||
				(this.subspecs.length && !this.initialSubspec) ||
				(this.subspecs.length === 1 &&
					this.subspecs[0].id !== this.initialSubspec.id);
		},
		filteredSubspecs() {
			if (this.subspecFilter && this.subspecs) {
				let subspecFilter = this.subspecFilter.toLowerCase();
				let filtered = [];
				for (var i = 0; i < this.subspecs.length; i++) {
					let o = this.subspecs[i];
					if (o.name.toLowerCase().indexOf(subspecFilter) >= 0 ||
						(o.desc && o.desc.toLowerCase().indexOf(subspecFilter) >= 0)) {
						filtered.push(o);
					}
				}
				return filtered;
			}
			return this.subspecs;
		},
		selectedSubspec() {
			if (this.subspecs && this.subspecId) {
				for (var i = 0; i < this.subspecs.length; i++) {
					if (this.subspecs[i].id === this.subspecId) {
						return this.subspecs[i];
					}
				}
			}
			return null;
		},
		disableCreateSubspec() {
			return !this.newSubspecName.trim() || this.creatingSubspec;
		},
		subspecNameMaxLength() {
			return window.const.specNameMaxLength;
		},

		// Output
		isValid() {
			switch (this.refType) {
				case REF_TYPE_URL:
					switch (this.urlMode) {
						case MODE_KEEP:
							return !!this.initialUrlObject;
						case MODE_CREATE:
							return isValidURL(this.newUrl);
						case MODE_SELECT:
							return !!this.urlId;
					}
					break;
				case REF_TYPE_SUBSPEC:
					switch (this.subspecMode) {
						case MODE_KEEP:
							return !!this.initialSubspec;
						case MODE_CREATE:
							return !!this.newSubspecName.trim();
						case MODE_SELECT:
							return !!this.subspecId;
					}
					break;
			}
			return false;
		},
		refFields() {
			// refType is passed separately to ajaxCreateBlock/ajaxSaveBlock
			switch (this.refType) {
				case REF_TYPE_URL:
					switch (this.urlMode) {
						case MODE_KEEP:
							if (this.initialUrlObject) {
								return {refId: this.initialUrlObject.id};
							}
							return null;
						case MODE_CREATE:
							return {refUrl: this.newUrl};
						case MODE_SELECT:
							return {refId: this.urlId};
					}
					break;
				case REF_TYPE_SUBSPEC:
					switch (this.subspecMode) {
						case MODE_KEEP:
							if (this.initialSubspec) {
								return {refId: this.initialSubspec.id};
							}
							return null;
						case MODE_CREATE:
							return {
								refName: this.newSubspecName,
								refDesc: this.newSubspecDesc,
							};
						case MODE_SELECT:
							return {refId: this.subspecId};
					}
					break;
			}
			return null;
		},

	},
	watch: {

		// URLs
		urlMode(mode) {
			if (mode === MODE_SELECT && !this.urlObjectsLoaded) {
				this.loadUrlObjects();
			}
		},

		// Subspecs
		subspecMode(mode) {
			if (mode === MODE_SELECT && !this.subspecsLoaded) {
				this.loadSubspecs();
			}
		},

		// Output
		isValid: {
			immediate: true,
			handler(valid) {
				// update sync prop value
				this.$emit('update:valid', valid);
			},
		},
		refFields: {
			immediate: true,
			deep: true,
			handler(fields) {
				// update sync prop value
				this.$emit('update:fields', fields);
			},
		},

	},
	methods: {

		// URLs
		loadUrlObjects() {
			if (this.loadingUrlObjects) {
				return;
			}
			this.loadingUrlObjects = true;
			ajaxLoadUrls(this.specId).then(urls => {
				this.urlObjects = urls;
				if (!urls.length) {
					// Enter create mode - none to select or keep
					this.urlMode = MODE_CREATE;
					notifyInfo('No URLs available.');
				}
				this.urlObjectsLoaded = true;
				this.loadingUrlObjects = false;
			}).fail(() => {
				if (!this.urlObjects) {
					// Nothing loaded or added
					if (this.initialUrlObject) {
						this.urlMode = MODE_KEEP;
					} else {
						this.urlMode = MODE_CREATE;
					}
				}
				this.urlObjectsLoaded = true;
				this.loadingUrlObjects = false;
			});
		},
		filterUrlObjects(filter) {
			// add a delay to filtering
			if (!filter) {
				this.urlFilter = '';
			} else {
				if (!this.debouncedFilterURLs) {
					this.debouncedFilterURLs = debounce(filter => {
						this.urlFilter = filter;
					});
				}
				this.debouncedFilterURLs(filter);
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
				if (this.urlId === urlObject.id) {
					this.urlId = null;
				}
			});
		},
		raisePlayVideo(urlObject) {
			this.$emit('play-video', urlObject);
		},
		createUrl() {
			// Create now - don't wait until save block
			this.creatingUrlObject = true;
			ajaxCreateUrl(this.specId, this.newUrl).then(urlObject => {
				this.urlId = urlObject.id;
				this.urlMode = MODE_SELECT;
				this.newUrl = '';
				this.creatingUrlObject = false;
				if (this.urlObjects) {
					// Check for existing
					for (let i = 0; i < this.urlObjects.length; i++) {
						if (this.urlObjects[i].id === urlObject.id) {
							return;
						}
					}
					// Add newly created URL
					this.urlObjects.push(urlObject);
				} else {
					this.urlObjects = [urlObject];
				}
				notifySuccess('URL created.');
			}).fail(() => {
				this.creatingUrlObject = false;
			});
		},

		// Subspecs
		loadSubspecs() {
			if (this.loadingSubspecs) {
				return;
			}
			this.loadingSubspecs = true;
			ajaxLoadSubspecs(this.specId).then(subspecs => {
				this.subspecs = subspecs;
				if (!subspecs.length) {
					// Enter create mode - none to select or keep
					this.subspecMode = MODE_CREATE;
					notifyInfo('No subspecs available.');
				}
				this.subspecsLoaded = true;
				this.loadingSubspecs = false;
			}).fail(() => {
				if (!this.subspecs) {
					// Nothing loaded or added
					if (this.initialSubspec) {
						this.subspecMode = MODE_KEEP;
					} else {
						this.subspecMode = MODE_CREATE;
					}
				}
				this.subspecsLoaded = true;
				this.loadingSubspecs = false;
			});
		},
		createSubspec() {
			// Create now - don't wait until save block
			this.creatingSubspec = true;
			ajaxCreateSubspec(this.specId,
				this.newSubspecName, this.newSubspecDesc,
			).then(subspec => {
				this.subspecId = subspec.id;
				this.subspecMode = MODE_SELECT;
				this.newSubspecName = '';
				this.newSubspecDesc = '';
				this.creatingSubspec = false;
				if (this.subspecs) {
					// Add newly created subspec
					this.subspecs.push(subspec);
				} else {
					this.subspecs = [subspec];
				}
				notifySuccess('Subspec created.');
			}).fail(() => {
				this.creatingSubspec = false;
			});
		},
		urlUpdated(urlObject) {
			if (this.urlObjects) {
				for (let i = 0; i < this.urlObjects.length; i++) {
					if (this.urlObjects[i].id === urlObject.id) {
						this.urlObjects[i] = urlObject;
						break;
					}
				}
			}
		},
	},
};
</script>

<style lang="scss">
.ref-form {
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
	.materian-icons {
		display: inline-block;
		margin-right: 1ex;
	}
}
</style>

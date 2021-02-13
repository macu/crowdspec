<template>
<el-dialog
	:title="(loading || block) ? 'Edit block' : 'Add block'"
	:visible.sync="showing"
	:width="$store.getters.dialogLargeWidth"
	:close-on-click-modal="false"
	@closed="closed()"
	class="spec-edit-block-modal">

	<p v-if="loading"><i class="el-icon-loading"/> Loading...</p>

	<template v-else-if="error">

		<p>{{error}}</p>

		<span slot="footer" class="dialog-footer">
			<el-button @click="showing = false">Close</el-button>
		</span>

	</template>

	<template v-else>

		<p v-if="block">
			Created <strong><moment :datetime="block.created"/></strong>;
			last modified <strong><moment :datetime="block.updated" :offset="true"/></strong>
		</p>

		<section>
			<el-radio-group v-model="styleType">
				<el-radio :label="STYLE_TYPE_BULLET">Bullet point</el-radio>
				<el-radio :label="STYLE_TYPE_NUMBERED">Numbered point</el-radio>
				<el-radio :label="STYLE_TYPE_NONE">Indented block</el-radio>
			</el-radio-group>

			<label>
				Title
				<el-input ref="titleInput" v-model="title" :maxlength="titleMaxLength" clearable/>
			</label>
		</section>

		<section>
			<el-radio-group v-model="refType">
				<el-radio :label="null">No media</el-radio>
				<el-radio :label="REF_TYPE_SUBSPEC">Subspec</el-radio>
				<el-radio :label="REF_TYPE_URL">URL</el-radio>
			</el-radio-group>

			<ref-url-form
				v-if="refType === REF_TYPE_URL"
				:spec-id="specId"
				:initial-url-object="existingUrlRefItem"
				:valid.sync="refFieldsValid"
				:fields.sync="refFields"
				@open-edit-url="openEditUrl"
				@play-video="raisePlayVideo"
				/>

			<ref-subspec-form
				v-else-if="refType === REF_TYPE_SUBSPEC"
				:spec-id="specId"
				:initial-subspec="existingSubspecRefItem"
				:valid.sync="refFieldsValid"
				:fields.sync="refFields"
				/>
		</section>

		<section>
			<el-radio-group v-model="contentType">
				<el-radio :label="CONTENT_TYPE_PLAIN">Plain text</el-radio>
				<el-radio :label="CONTENT_TYPE_MARKDOWN">Markdown</el-radio>
			</el-radio-group>

			<template v-if="contentType === CONTENT_TYPE_MARKDOWN">
				<div class="split-even">
					<label>
						Body
						<el-input type="textarea" v-model="body" :autosize="{minRows: 2}"/>
					</label>
					<label>
						Preview
						<div v-if="previewError" class="error">{{previewError}}</div>
						<div v-else class="markdown" v-html="previewHtml" v-loading="loadingPreview"/>
					</label>
				</div>
			</template>
			<label v-else>
				Body
				<el-input type="textarea" v-model="body" :autosize="{minRows: 2}"/>
			</label>
		</section>

		<span slot="footer" class="dialog-footer">
			<el-button @click="showing = false">Cancel</el-button>
			<el-button v-if="block" @click="promotDelete()" type="warning">Delete</el-button>
			<el-button @click="submit()" type="primary">
				{{block ? 'Save' : 'Add'}}
			</el-button>
		</span>

	</template>

</el-dialog>
</template>

<script>
import $ from 'jquery';
import Moment from '../widgets/moment.vue';
import RefUrlForm from './ref-url-form.vue';
import RefSubspecForm from './ref-subspec-form.vue';
import {
	ajaxLoadBlockForEditing, ajaxCreateBlock, ajaxSaveBlock,
	ajaxRenderMarkdown,
} from './ajax.js';
import {
	STYLE_TYPE_NONE, STYLE_TYPE_BULLET, STYLE_TYPE_NUMBERED,
	CONTENT_TYPE_PLAIN, CONTENT_TYPE_MARKDOWN,
	REF_TYPE_URL, REF_TYPE_SUBSPEC,
} from './const.js';
import {
	debounce,
} from '../utils.js';

export default {
	components: {
		Moment,
		RefUrlForm,
		RefSubspecForm,
	},
	props: {
		specId: {
			type: Number,
			required: true,
		},
		subspecId: Number,
	},
	data() {
		return {
			// user inputs
			contentType: CONTENT_TYPE_PLAIN,
			styleType: STYLE_TYPE_BULLET,
			title: '',
			body: '',
			refType: null,
			refFields: null,
			refFieldsValid: false,
			// passed in
			block: null,
			parentId: null,
			insertBeforeId: null,
			callback: null,
			// state
			showing: false,
			loading: false,
			error: null,
			loadingPreview: false,
			previewHtml: '',
			previewXhr: null, // pending render request
			previewError: null,
		};
	},
	computed: {
		STYLE_TYPE_NONE() {
			return STYLE_TYPE_NONE;
		},
		STYLE_TYPE_BULLET() {
			return STYLE_TYPE_BULLET;
		},
		STYLE_TYPE_NUMBERED() {
			return STYLE_TYPE_NUMBERED;
		},
		CONTENT_TYPE_PLAIN() {
			return CONTENT_TYPE_PLAIN;
		},
		CONTENT_TYPE_MARKDOWN() {
			return CONTENT_TYPE_MARKDOWN;
		},
		REF_TYPE_URL() {
			return REF_TYPE_URL;
		},
		REF_TYPE_SUBSPEC() {
			return REF_TYPE_SUBSPEC;
		},
		titleMaxLength() {
			return window.const.blockTitleMaxLength;
		},
		existingUrlRefItem() {
			return this.block && this.block.refType === REF_TYPE_URL && this.block.refItem || null;
		},
		existingSubspecRefItem() {
			return this.block && this.block.refType === REF_TYPE_SUBSPEC && this.block.refItem || null;
		},
	},
	watch: {
		refType() {
			this.refFields = null;
		},
		body() {
			this.renderPreview();
		},
		contentType() {
			this.renderPreview();
		},
	},
	methods: {
		renderPreview() {
			if (this.contentType === CONTENT_TYPE_MARKDOWN) {
				if (this.body.trim()) {
					if (!this.debouncedBodyRender) {
						this.debouncedBodyRender = debounce(content => {
							this.loadingPreview = true;
							this.previewError = null;

							if (this.previewXhr) {
								// Cancel previous call
								this.previewXhr.abort();
							}

							let xhr = ajaxRenderMarkdown(content);

							// Store request so only latest request gets resolved
							this.previewXhr = xhr;

							xhr.then(response => {
								// Only resolve most recent call
								if (this.previewXhr === xhr) {
									if (response.error) {
										this.previewError = response.error;
									} else {
										this.previewHtml = response.html;
									}
									this.loadingPreview = false;
									this.previewXhr = null;
								}
							}).fail(() => {
								// Only resolve most recent call
								if (this.previewXhr === xhr) {
									this.previewError = 'Error generating preview.';
									this.loadingPreview = false;
									this.previewXhr = null;
								}
							});
						}, 1000);
					}
					this.debouncedBodyRender(this.body);
				} else {
					this.previewHtml = '';
				}
			}
		},
		showAdd(parentId, insertBeforeId, callback) {
			this.parentId = parentId;
			this.insertBeforeId = insertBeforeId;
			this.callback = callback;
			this.showing = true;
			this.focusTitleInput();
		},
		showEdit(blockId, callback) {
			this.loading = true;
			this.error = null;
			this.showing = true;
			ajaxLoadBlockForEditing(this.specId, blockId).then(block => {
				this.loading = false;
				this.block = block;
				this.styleType = block.styleType;
				this.contentType = block.contentType;
				this.title = block.title || '';
				this.body = block.body || '';
				this.refType = block.refType;
				this.callback = callback;
				this.focusTitleInput();
			}).fail(() => {
				this.error = 'Error loading block';
				this.loading = false;
			});
		},
		focusTitleInput() {
			this.$nextTick(() => {
				$('input', this.$refs.titleInput.$el).focus();
			});
		},
		openEditUrl(urlObject, updated = null, deleted = null) {
			this.$emit('open-edit-url', urlObject, updatedUrlObject => {
				// Updated
				if (this.existingUrlRefItem && updatedUrlObject.id === this.existingUrlRefItem.id) {
					// Update existing ref
					this.block.refItem = updatedUrlObject;
				}
				if (updated) {
					updated(updatedUrlObject);
				}
			}, deletedId => {
				// Deleted
				if (this.existingUrlRefItem && deletedId === this.existingUrlRefItem.id) {
					// Clear existing ref
					this.block.refType = null;
					this.block.refId = null;
					this.block.refItem = null;
				}
				if (deleted) {
					deleted(deletedId);
				}
			});
		},
		raisePlayVideo(urlObject) {
			this.$emit('play-video', urlObject);
		},
		promotDelete() {
			this.$emit('prompt-delete', this.block.id, () => {
				this.showing = false;
			});
		},
		submit() {
			if (this.refType) {
				if (!(this.refFieldsValid && this.refFields)) {
					this.$alert('Please specify the attachment or choose "No media".', {
						type: 'error',
					});
					return;
				}
			}
			if (this.contentType === CONTENT_TYPE_MARKDOWN && this.previewError) {
				this.$alert('Error rendering preview. Please check your HTML syntax.', {
					type: 'error',
				});
				return;
			}
			if (!(this.title.trim() || this.body.trim())) {
				this.$alert('A title, body, or attachment is required.', {
					type: 'error',
				});
				return;
			}
			let sending = this.createSendingSpinner();
			let callback = this.callback; // in case modal is closed before complete
			if (this.block) {
				// TODO only send title and body if changed
				ajaxSaveBlock(
					this.specId,
					this.block.id,
					this.styleType,
					this.contentType,
					this.title,
					this.body,
					this.refType,
					this.refFields,
				).then(updatedBlock => {
					callback(updatedBlock);
					this.showing = false;
					sending.close();
				}).fail(() => {
					sending.close();
				});
			} else {
				ajaxCreateBlock(
					this.specId,
					this.subspecId,
					this.parentId,
					this.insertBeforeId,
					this.styleType,
					this.contentType,
					this.title,
					this.body,
					this.refType,
					this.refFields,
				).then(newBlock => {
					callback(newBlock);
					this.showing = false;
					sending.close();
				}).fail(() => {
					sending.close();
				});
			}
		},
		createSendingSpinner() {
			return this.$loading({
				lock: true,
				background: 'rgba(0, 0, 0, 0.7)',
			});
		},
		closed() {
			this.block = null;
			this.error = null;
			this.parentId = null;
			this.insertBeforeId = null;
			this.callback = null;
			// leave styleTyle and contentType set to the last values to appear
			// TODO initialize upon modal open according to sibling blocks
			this.title = '';
			this.body = '';
			this.refType = null;
			this.refFields = null;
			this.refFieldsValid = false;
			this.previewHtml = '';
			this.previewError = null;
		},
	},
};
</script>

<style lang="scss">
@import '../_styles/_breakpoints.scss';
@import '../_styles/_colours.scss';

.spec-edit-block-modal {
	>.el-dialog {
		>.el-dialog__body {
			>p {
				margin-top: 0;
			}
			>section {
				margin-top: 40px;
				>*+* {
					margin-top: 20px;
				}
				>label {
					display: block;
					>.el-input {
						display: block;
						width: 100%;
					}
				}
				>.el-radio-group {
					display: block;
					padding: 10px;
					background-color: $section-highlight;
				}
				>.split-even {
					display: grid;
					grid-template-columns: calc(50% - 10px) calc(50% - 10px);
					column-gap: 20px;
					row-gap: 20px;
					>label {
						display: block;
					}
					@include mobile {
						grid-template-columns: 100%;
					}
					.markdown, .error {
						padding: 15px;
						border-radius: 4px;
						min-height: 54px;
					}
					.markdown {
						border: thin solid lightgreen;
					}
					.error {
						border: thin solid red;
					}
				}
			}
		}
	}
}
</style>

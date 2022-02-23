<template>
<el-dialog
	:title="(loading || block) ? 'Edit block' : 'Add block'"
	v-model="showing"
	:width="$store.getters.dialogLargeWidth"
	:close-on-click-modal="false"
	@closed="closed()"
	custom-class="spec-edit-block-modal">

	<p v-if="loading">
		<loading-message message="Loading..."/>
	</p>

	<template v-else-if="error">

		<p>{{error}}</p>

	</template>

	<template v-else-if="showing">

		<p v-if="block">
			Created <strong><moment :datetime="block.created"/></strong>;
			last modified <strong><moment :datetime="block.updated" :offset="true"/></strong>
		</p>

		<section>
			<div class="field">
				<el-radio-group v-model="styleType">
					<el-radio :label="STYLE_TYPE_BULLET">Bullet point</el-radio>
					<el-radio :label="STYLE_TYPE_NUMBERED">Numbered point</el-radio>
					<el-radio :label="STYLE_TYPE_NONE">Indented block</el-radio>
				</el-radio-group>
			</div>

			<label>
				<div>Title</div>
				<el-input ref="titleInput" v-model="title" :maxlength="titleMaxLength" clearable/>
			</label>
		</section>

		<section>
			<div class="field">
				<el-radio-group v-model="refType">
					<el-radio :label="null">No media</el-radio>
					<el-radio :label="REF_TYPE_SUBSPEC">Subspec</el-radio>
					<el-radio :label="REF_TYPE_URL">URL</el-radio>
				</el-radio-group>
			</div>

			<ref-form
				ref="refForm"
				v-show="!!refType"
				:spec-id="specId"
				:ref-type="refType"
				:existing-ref-type="existingRefType"
				:existing-ref-item="existingRefItem"
				v-model:fields="refFields"
				v-model:valid="refFieldsValid"
				@open-edit-url="openEditUrl"
				@play-video="raisePlayVideo"
				/>
		</section>

		<section>
			<div class="field">
				<el-radio-group v-model="contentType">
					<el-radio :label="CONTENT_TYPE_PLAIN">Plain text</el-radio>
					<el-radio :label="CONTENT_TYPE_MARKDOWN">Markdown</el-radio>
				</el-radio-group>
			</div>

			<template v-if="contentType === CONTENT_TYPE_MARKDOWN">
				<div class="split-even">
					<!-- display preview beneath for final review-->
					<label>
						<div>Body</div>
						<el-input type="textarea" v-model="body" :autosize="{minRows: 2}"/>
					</label>
					<label>
						<div>Preview</div>
						<div v-if="previewError" class="error">{{previewError}}</div>
						<div v-else ref="renderedHtml" class="markdown" v-html="previewHtml" v-loading="loadingPreview"/>
					</label>
				</div>
			</template>
			<label v-else>
				<div>Body</div>
				<el-input type="textarea" v-model="body" :autosize="{minRows: 2}"/>
			</label>
		</section>

	</template>

	<template #footer>

		<span v-if="error" class="dialog-footer">
			<el-button @click="showing = false">Close</el-button>
		</span>

		<span v-else-if="showing" class="dialog-footer">
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
import RefForm from './ref-form.vue';
import LoadingMessage from '../widgets/loading.vue';
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
import {
	SCRIPT_HLJS,
	loadScript,
} from '../widgets/script-loader.js';

export default {
	components: {
		Moment,
		RefForm,
		LoadingMessage,
	},
	props: {
		specId: {
			type: Number,
			required: true,
		},
		subspecId: Number,
	},
	emits: ['open-edit-url', 'play-video', 'prompt-delete'],
	data() {
		return {
			// user inputs
			styleType: STYLE_TYPE_BULLET,
			contentType: CONTENT_TYPE_PLAIN,
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
		existingRefType() {
			return this.block && this.block.refType;
		},
		existingRefItem() {
			return this.block && this.block.refType && this.block.refItem || null;
		},
		initialUrlObject() {
			return this.existingRefType === REF_TYPE_URL && this.existingRefItem;
		},
	},
	watch: {
		body() {
			this.renderMarkdownPreview();
		},
		contentType() {
			this.renderMarkdownPreview();
		},
		previewHtml(html) {
			this.addCodeHighlighting();
		},
	},
	methods: {
		showAddBlock(parentId, insertBeforeId, defaultStyleType, callback) {
			if (defaultStyleType === true) {
				// not all these routes are followed;
				// defaultStyleType is given as a string when adding above or below a block,
				// which overrides the behaviour defined here;
				// but the logic here is comprehensive as a fallback in case I want to use it
				if (insertBeforeId) {
					// Add before an existing block
					let $sibling = $('[data-spec-block="'+insertBeforeId+'"]', '.spec-view');
					let $prevSibling = $sibling.prev('[data-spec-block]');
					if ($prevSibling.length) {
						// Copy style of preceeding block
						defaultStyleType = $prevSibling.data('vc').getStyleType();
					} else {
						// Copy style of following block
						defaultStyleType = $sibling.data('vc').getStyleType();
					}
				} else if (parentId) {
					// Add at last position within parent block
					let $parent = $('[data-spec-block="'+parentId+'"]', '.spec-view');
					let $sibling = $('>ul>[data-spec-block]:last-child', $parent);
					if ($sibling.length) {
						// Copy style of last block currently in parent
						defaultStyleType = $sibling.data('vc').getStyleType();
					}
				} else {
					let $sibling = $('>ul>[data-spec-block]:last-child', '.spec-view');
					if ($sibling.length) {
						// Copy style of current last root block
						defaultStyleType = $sibling.data('vc').getStyleType();
					}
				}
			}
			if (typeof defaultStyleType === 'string') {
				this.styleType = defaultStyleType;
			}
			this.parentId = parentId;
			this.insertBeforeId = insertBeforeId;
			this.callback = callback;
			this.showing = true;
			this.focusTitleInput();
		},
		showEditBlock(blockId, callback) {
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
		renderMarkdownPreview() {
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
		addCodeHighlighting() {
			if (!this.previewHtml) {
				return;
			}
			this.$nextTick(() => {
				let $codeblocks = $('pre>code[class*="language-"]', this.$refs.renderedHtml);
				if ($codeblocks.length) {
					loadScript(SCRIPT_HLJS).then(hljs => {
						$codeblocks.each((i, e) => {
							hljs.highlightBlock(e);
						});
					});
				}
			});
		},
		openEditUrl(urlObject, updated = null, deleted = null) {
			this.$emit('open-edit-url', urlObject, updatedUrlObject => {
				// Updated
				if (this.initialUrlObject && updatedUrlObject.id === this.initialUrlObject.id) {
					// Update existing ref
					this.block.refItem = updatedUrlObject;
				}
				this.$refs.refForm.urlUpdated(urlObject);
				if (updated) {
					updated(updatedUrlObject);
				}
			}, deletedId => {
				// Deleted
				if (this.initialUrlObject && deletedId === this.initialUrlObject.id) {
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
			} else if (!this.refType && !(this.title.trim() || this.body.trim())) {
				this.$alert('A title, body, or attachment is required.', {
					type: 'error',
				});
				return;
			}
			if (this.contentType === CONTENT_TYPE_MARKDOWN && this.previewError) {
				this.$alert('Error rendering preview. Please check your HTML syntax.', {
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
			this.styleType = STYLE_TYPE_BULLET;
			this.contentType = CONTENT_TYPE_PLAIN;
			this.block = null;
			this.error = null;
			this.parentId = null;
			this.insertBeforeId = null;
			this.callback = null;
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

.spec-edit-block-modal.el-dialog {
	>.el-dialog__body {
		>p {
			margin-top: 0;
		}
		>section {
			&:not(:first-child) {
				margin-top: 40px;
			}
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
			>.field {
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
</style>

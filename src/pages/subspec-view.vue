<template>
<div class="subspec-view-page" :class="{loading}">
	<p v-if="loading">
		<loading-message message="Loading..."/>
	</p>
	<spec-view
		v-else-if="subspec"
		:key="subspec.renderTime"
		:subspec="subspec"
		:enable-editing="enableEditing"
		@prompt-nav-spec="promptNavSpec"
		@open-community="openCommunity"
		@play-video="playVideo"
		@rendered="$emit('rendered')"
		/>
	<p v-else>
		Failed to load subspec.
	</p>
</div>
</template>

<script>
import SpecView from '../spec/view.vue';
import LoadingMessage from '../widgets/loading.vue';

export default {
	components: {
		SpecView,
		LoadingMessage,
	},
	props: {
		loading: Boolean,
		subspec: Object,
		enableEditing: Boolean,
	},
	emits: ['rendered', 'prompt-nav-spec', 'open-community', 'play-video'],
	methods: {
		promptNavSpec() {
			this.$emit('prompt-nav-spec');
		},
		openCommunity(targetType, targetId, onAdjustUnread, onAdjustComments) {
			this.$emit('open-community', targetType, targetId, onAdjustUnread, onAdjustComments);
		},
		playVideo(urlObject) {
			this.$emit('play-video', urlObject);
		},
	},
};
</script>

<style lang="scss">
@use "sass:math";
@import '../_styles/_app.scss';
@import '../_styles/_breakpoints.scss';

.subspec-view-page {
	padding: $content-area-padding;
	padding-top: math.div($content-area-padding, 2);

	@include mobile {
		padding: $content-area-padding-sm 0;

		&.loading {
			padding: $content-area-padding-sm;
		}
	}
}
</style>

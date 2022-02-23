<template>
<el-dialog
	v-model="showing"
	:width="$store.getters.dialogSmallWidth"
	@closed="closed()"
	custom-class="play-video-modal">

	<div v-if="vid" class="embed-responsive">

		<iframe v-if="vid.type === 'youtube'"
			:src="'https://www.youtube.com/embed/'+vid.id+'?rel=0&amp;showinfo=0&amp;ecver=1&amp;autoplay=1'"
			webkitallowfullscreen mozallowfullscreen allowfullscreen></iframe>

		<iframe v-else-if="vid.type === 'vimeo'"
			:src="'https://player.vimeo.com/video/'+vid.id+'?title=0&amp;byline=0&amp;portrait=0&amp;autoplay=1'"
			webkitallowfullscreen mozallowfullscreen allowfullscreen></iframe>

	</div>

</el-dialog>
</template>

<script>
import {extractVid} from '../utils.js';

export default {
	data() {
		return {
			vid: null,
			showing: false,
		};
	},
	methods: {
		show(urlObject) {
			let vid = extractVid(urlObject.url);
			if (vid) {
				this.vid = vid;
				this.showing = true;
			}
		},
		closed() {
			this.vid = null;
		},
	},
};
</script>

<style lang="scss">
.play-video-modal.el-dialog {
	>.el-dialog__body {
		>.embed-responsive {
			// adapted from Bootstrap embed-responsive embed-responsive-16by9
			position: relative;
			display: block;
			width: 100%;
			padding: 0;
			overflow: hidden;
			&::before {
				display: block;
				content: "";
				padding-top: 56.25%; // 16x9
			}
			>iframe {
				position: absolute;
				top: 0;
				bottom: 0;
				left: 0;
				width: 100%;
				height: 100%;
				border: 0;
			}
		}
	}
}
</style>

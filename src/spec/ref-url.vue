<template>
<div class="ref-url ref-item">
	<img v-if="item.imageData" :src="item.imageData"/>
	<div>
		<div v-if="showEdit || showPlay" class="right">
			<el-button v-if="showEdit" @click="$emit('edit')" size="small" circle>
				<i class="material-icons">edit</i>
			</el-button>
			<el-button v-if="showPlay" @click="$emit('play')" size="small" type="primary" class="play-button">Play</el-button>
		</div>
		<div class="title">
			<span><i class="material-icons">link</i> {{showPlay ? 'Video' : 'URL'}}</span>
			<a :href="item.url" target="_blank" @click.stop>{{(item.title && item.title.trim()) || item.url}}</a>
		</div>
		<div v-if="item.desc" class="desc">{{item.desc.trim()}}</div>
	</div>
</div>
</template>

<script>
import {isVideoURL} from '../utils.js';

export default {
	props: {
		item: Object,
		showEdit: Boolean,
	},
	emits: ['edit', 'play'],
	computed: {
		showPlay() {
			return isVideoURL(this.item.url);
		},
	},
};
</script>

<style lang="scss">
.ref-url {
	border: thin solid lightgray;
	border-radius: 5px;
	padding: 10px;
	display: flex;
	align-items: flex-start;
	flex-direction: row;
	>img {
		width: 50px;
		margin-right: 10px;
	}
	>div {
		flex: 1;
		>.right {
			float: right;
			margin-left: 10px;
			>.el-button {
				&:not(.play-button) {
					padding: 3px;
				}
				&.play-button {
					padding-top: 3px;
					padding-bottom: 3px;
				}
				font-size: 12px;
			}
			>.el-button+.el-button {
				margin-left: 5px;
			}
		}
		>.title {
			font-weight: bold;
			overflow-wrap: break-word; // IE
			overflow-wrap: anywhere;
			>span {
				font-weight: normal;
				color: gray;
				display: inline-block;
				margin-right: 5px;
			}
		}
		>.desc {
			margin-top: 5px;
		}
	}
}
</style>

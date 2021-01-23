<template>
	<span class="tag">
		<span v-if="adminPinned" class="part"><i class="el-icon-success"></i></span>
		<span v-if="authorClaimed" class="part"><i class="el-icon-user"></i></span>
		<span class="part">{{tag.name}}</span>
		<el-tooltip placement="bottom">
			<span class="part">{{voteSum}}</span>
			<div slot="content">
				<div><i class="el-icon-plus"></i> {{assentVotes}}</div>
				<div><i class="el-icon-minus"></i> {{dissentVotes}}</div>
			</div>
		</el-tooltip>
		<span class="btn assent"
			:class="{'user-vote': hasUserAssent, sending}"
			@click="toggleAssent()"><i class="el-icon-plus"></i></span>
		<span class="btn dissent"
			:class="{'user-vote': hasUserDissent, sending}"
			@click="toggleDissent()"><i class="el-icon-minus"></i></span>
		<span v-if="userIsAdmin" class="btn"
			@click="openOptions()"><i class="el-icon-setting"></i></span>
	</span>
</template>

<script>
import $ from 'jquery';
import {alertError} from '../utils.js';

export default {
	props: {
		tag: Object,
		userIsAdmin: Boolean,
	},
	data() {
		adminPinned: this.tag.pinned || false,
		userVote: this.tag.userVote || null,
		sending: false,
	},
	computed: {
		authorClaimed() {
			return this.tag.authorClaimed || false;
		},
		voteSum() {
			if (this.userVote === 'assent') {
				return tag.sum + 1;
			} else if (this.userVote === 'dissent') {
				return tag.sum - 1;
			}
			return tag.sum;
		},
		hasUserAssent() {
			return this.userVote === 'assent';
		},
		hasUserDissent() {
			return this.userVote === 'dissent';
		},
		assentVotes() {
			if (this.hasUserAssent) {
				return this.tag.assentVotes + 1;
			}
			return this.tag.assentVotes;
		},
		dissentVotes() {
			if (this.hasUserDissent) {
				return this.tag.dissentVotes + 1;
			}
			return this.tag.dissentVotes;
		},
		voteParams() {
			return {
				tagId: this.tag.id,
				specId: this.tag.specId,
				subspecId: this.tag.subspecId,
				commentId: this.tag.commentId,
			};
		},
	},
	methods: {
		toggleAssent() {
			this.sending = true;
			let userVote = this.hasUserAssent() ? null : 'assent';
			$.post('', $.extend({userVote}, this.voteParams)).then(() => {
				this.sending = false;
				this.userVote = userVote === 'assent' ? 'assent' : null;
			}).fail(jqXHR => {
				this.sending = false;
				alertError(jqXHR);
			});
		},
		toggleDissent() {
			this.sending = true;
			let userVote = this.hasUserDissent() ? null : 'dissent';
			$.post('', $.extend({userVote}, this.voteParams)).then(() => {
				this.sending = false;
				this.userVote = userVote === 'dissent' ? 'dissent' : null;
			}).fail(jqXHR => {
				this.sending = false;
				alertError(jqXHR);
			});
		},
		openOptions() {
			this.$emit('open-options', this.tag, updatedTag => {
				this.adminPinned = updatedTag.adminPinned;
			}, () => {
				this.$emit('delete-tag', this.tag.id);
			});
		},
	},
};
</script>

<style lang="scss">
.tag {
	display: flex;
	flex-direction: row;
	line-height: 24px;
	font-size: 12px;
	>.part, >.btn {
		display: inline-block;
		padding: 0 10px;
		color: #409eff;
		background-color: #ecf5ff;
		white-space: nowrap;
	}
	>.btn {
		cursor: pointer;
		&:hover {
			color: white;
			background-color: #409eff;
		}
		&.assent.user-vote {
			color: darken(green, 50%);
			background-color: lighten(green, 50%);
		}
		&.dissent.user-vote {
			color: darken(red, 50%);
			background-color: lighten(red, 50%);
		}
		&.sending {
			cursor: progress;
			color: lighten(#409eff, 20%);
			background-color: darken(#ecf5ff, 20%);
			&.assent.user-vote {
				color: darken(green, 30%);
				background-color: lighten(green, 30%);
			}
			&.dissent.user-vote {
				color: darken(red, 30%);
				background-color: lighten(red, 30%);
			}
		}
	}
	>*:not(:first-child) {
		border-left: 1px solid #d9ecff;
		&.assent.user-vote {
			border-left: 1px solid green;
		}
		&.dissent.user-vote {
			border-left: 1px solid red;
		}
	}
	>*:first-child {
		border-top-left-radius: 4px;
		border-bottom-left-radius: 4px;
	}
	>*:last-child {
		border-top-right-radius: 4px;
		border-bottom-right-radius: 4px;
	}
}
</style>

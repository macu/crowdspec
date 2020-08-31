<template>
<span>{{output}}</span>
</template>

<script>
import moment from 'moment';

export default {
	props: {
		datetime: String,
		offset: Boolean,
	},
	data() {
		return {
			timeout: null,
		};
	},
	computed: {
		moment() {
			return moment.parseZone(this.datetime);
		},
		currentTime() {
			return this.$store.state.currentTime;
		},
		output() {
			// Add dependency on currentTime to continually recompute this prop
			if (this.currentTime > 0) {
				if (this.offset) {
					return this.moment.fromNow();
				} else {
					return this.moment.format('YYYY-MM-DD h:mm a');
				}
			}
			return null;
		},
	},
};
</script>

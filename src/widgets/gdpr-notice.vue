<template>
	<div>
		<p>Your cache is clear.</p>
		<el-button v-if="hideInfo">[About tracking]</el-button>
		<el-alert v-else>
			Google Analytics is used to collect pageview and visitor statistics.

		</el-alert>

	[Decline]
	[Enable visitor tracking cookie]
	[Decline cookie and local storage]
	</div>
</template>

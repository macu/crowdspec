<template>
<div class="app">

	<header>
		App
		<!-- <h1 @click="gotoIndex()">CrowdSpec</h1> -->
		<div>
			<!-- <el-button @click="openEditProfile()" type="text" size="mini" class="username-button">
				<username :username="username" :highlight="highlight"/>
			</el-button> -->
			<!-- <el-button @click="logout()" size="mini">Log out</el-button> -->
		</div>
	</header>

</div>
</template>

<script>
export default {
	methods: {
		gotoIndex() {

		},
		logout() {

		},
	}
};
</script>

<style lang="scss">
.app {
	color: blue;

	>header {
		display: flex;
		flex-direction: row; // horizontal
		align-items: center; // vertical align
		justify-content: flex-end; // align right
		flex-wrap: wrap;
		background-color: blue;
		color: white;
	}

}
</style>

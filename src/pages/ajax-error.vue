<template>
<div>
	<template v-if="parseInt($route.params.code, 10) === 0">
		Could not connect to server
	</template>
	<template v-else>
		Request failed with error code {{$route.params.code}}
	</template>
</div>
</template>

<script>
export default {
};
</script>

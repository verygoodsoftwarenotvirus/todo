<template>
	<div class="container">
		<ItemCreator />
	</div>
</template>

<script lang='ts'>
import ItemCreator from "@/components/ItemCreator.vue"; // @ is an alias to /src
import { Component, Vue } from "vue-property-decorator";

@Component({
	components: {
		ItemCreator,
	},
})
export default class CreateItem extends Vue { }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>

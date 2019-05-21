<template>
	<div class="home">
		<HomePage msg="home page yay" />
	</div>
</template>

<script lang='ts'>
import HomePage from "@/components/HomePage.vue"; // @ is an alias to /src
import { Component, Vue } from "vue-property-decorator";

@Component({
	components: {
		HomePage,
	},
})
export default class Items extends Vue { }
</script>


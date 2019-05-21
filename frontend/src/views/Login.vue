<template>
	<div class="container">
		<LoginPage />
	</div>
</template>

<script lang='ts'>
import LoginPage from "@/components/LoginPage.vue"; // @ is an alias to /src
import { Component, Vue } from "vue-property-decorator";

@Component({
	components: {
		LoginPage,
	},
})

export default class Login extends Vue { }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>

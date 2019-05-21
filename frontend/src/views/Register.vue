<template>
	<div class="container">
		<RegistrationPage />
	</div>
</template>

<script lang='ts'>
import RegistrationPage from "@/components/RegistrationPage.vue"; // @ is an alias to /src
import { Component, Vue } from "vue-property-decorator";

@Component({
	components: {
		RegistrationPage,
	},
})
export default class Register extends Vue { }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>

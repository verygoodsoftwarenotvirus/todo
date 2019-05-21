<template>
	<div>
		<!-- use the modal component, pass in the prop -->
		<div v-if="twoFactorQRCode != ''" @close="showModal = false">
			<form @submit.prevent="goHome">
				<img height="200" width="200" v-bind:src="twoFactorQRCode" />
				<hr />
				<b-input type="submit" value="Got it!"></b-input>
			</form>
		</div>

		<form v-if="twoFactorQRCode == ''" @submit.prevent="createUser">
			<b-field label="Username">
				<b-input icon="account" v-model="username"></b-input>
			</b-field>

			<b-field label="Password">
				<b-input
					type="password"
					v-model="password"
					password-reveal
				></b-input>
			</b-field>

			<b-input type="submit" value="Register"></b-input>
		</form>
	</div>
</template>

<script lang="ts">
import axios, { AxiosResponse } from "axios";
import { Component, Prop, Vue } from "vue-property-decorator";

@Component
export default class RegistrationPage extends Vue {

	private username: string = "";
	private password: string = "";
	private twoFactorQRCode: string = "";
	private showModal: boolean = false;

	private createUser() {
		axios.post(
			"http://localhost:8080/users",
			{
				username: this.username,
				password: this.password,
			},
		).then((res: AxiosResponse) => {
			console.log("response received!");
			console.log(res.data);

			this.twoFactorQRCode = res.data.qr_code || "";
			console.log("2FA secret: ", res.data.two_factor_secret || "");
		}).catch(console.error);
	}

	private goHome() {
		this.$router.push("/");
	}
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>




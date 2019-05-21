<template>
	<form @submit.prevent="handleSubmit">
		<b-field label="Username">
			<b-input v-model="username" icon="account"></b-input>
		</b-field>

		<b-field label="Password">
			<b-input
				v-model="password"
				type="password"
				password-reveal
			></b-input>
		</b-field>

		<b-field label="2FA Code">
			<b-input v-model="totpToken" type="text"></b-input>
		</b-field>

		<b-input type="submit" value="Login" />
	</form>
</template>

<script lang="ts">
import axios, { AxiosRequestConfig, AxiosResponse } from "axios";
import { Component, /* Prop, */ Vue } from "vue-property-decorator";

@Component
export default class LoginPage extends Vue {

	private username: string = "";
	private password: string = "";
	private totpToken: string = "";

	private handleSubmit() {
		console.log(`sending this to login server: ${window.location.origin}/users/login`);
		console.log(
			{
				username: this.username,
				password: this.password,
				totp_token: this.totpToken,
			},
		);



		axios({
			method: "post",
			url: `${window.location.origin}/users/login`,
			data: {
				username: this.username,
				password: this.password,
				totp_token: this.totpToken,
			},
		}).then((res: AxiosResponse) => {
			this.$router.push("/");
		}).catch((reason) => {
			console.error(JSON.stringify(reason));
		});
	}

}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>




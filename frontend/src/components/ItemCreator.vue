<template>
	<form @submit.prevent="createItem">
		<b-field label="Name">
			<b-input v-model="name" type="text" maxlength="64"></b-input>
		</b-field>

		<b-field label="Details">
			<b-input
				v-model="details"
				type="textarea"
				maxlength="512"
			></b-input>
		</b-field>

		<b-input
			type="submit"
			:disabled="name !== '' && details !== ''"
			value="Save"
		/>
	</form>
</template>

<script lang="ts">
import axios, { AxiosRequestConfig, AxiosResponse } from "axios";
import { Vue } from "vue-property-decorator";

export default class ItemCreator extends Vue {

	public name: string = "";
	public details: string = "";

	public createItem() {
		axios.post(
			`${window.location.origin}/api/v1/items`,
			{
				name: this.name,
				details: this.details,
			},
		).then((res: AxiosResponse) => {
			this.name = "";
			this.details = "";

			console.log(`name: ${this.name} and details: ${this.details}`);
		}).catch(console.error);
	}

}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>




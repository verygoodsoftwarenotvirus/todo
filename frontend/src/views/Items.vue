<template>
	<div class="container">
		<div v-if="itemsData.length === 0">
			No data found
		</div>

		<div v-if="itemsData.length !== 0">
			<button
				class="button field is-danger"
				@click="deleteItems"
				:disabled="!checkedRows.length"
			>
				<b-icon pack="fas" icon="fa-trash"></b-icon>
				<span>Delete checked</span>
			</button>

			<div v-if="itemsData.length !== 0" class="container">
				<b-table
					:checked-rows.sync="checkedRows"
					:data="itemsData"
					:checkable="true"
					:columns="columns"
				></b-table>
			</div>
		</div>
	</div>
</template>

<script lang='ts'>
import Item from "@/service_client/models"; // @ is an alias to /src
import axios, { AxiosRequestConfig, AxiosResponse } from "axios";
import { Component, Vue } from "vue-property-decorator";

@Component
export default class ItemsPage extends Vue {
	public checkedRows: Item[] = [];
	public itemsData: Item[] = [];

	public data() {
		return {
			columns: [
				{
					field: "id",
					label: "ID",
					sortable: true,
					centered: true,
					numeric: true,
				},
				{
					field: "name",
					sortable: true,
					centered: true,
					label: "Name",
				},
				{
					field: "details",
					label: "Details",
				},
				{
					field: "updated_on",
					label: "Updated On",
					sortable: true,
					centered: true,
				},
				{
					field: "created_on",
					label: "Created On",
					sortable: true,
					centered: true,
				},
			],
		};
	}

	public mounted() {
		axios.get(`${window.location.origin}/api/v1/items`)
			.then((res) => {
				this.itemsData = res.data.items || [];
			}, console.error);
	}

	public deleteItems() {
		for (const i of this.checkedRows) {
			axios.delete(`${window.location.origin}/api/v1/items/${i.id}`)
				.then((res) => {
					const index: number = this.itemsData.indexOf(i, 0);
					if (index > -1) {
						this.itemsData.splice(index, 1);
					}
				}, console.error);
		}
		this.checkedRows = [];
	}

}
</script>



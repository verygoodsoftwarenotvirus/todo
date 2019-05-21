import Vue from "vue";
import Router from "vue-router";
import CreateItem from "./views/CreateItem.vue";
import Home from "./views/Home.vue";
import ItemsPage from "./views/Items.vue";
import Login from "./views/Login.vue";
import Register from "./views/Register.vue";

Vue.use(Router);

export default new Router({
	mode: "history",
	base: process.env.BASE_URL,
	routes: [
		{
			path: "/",
			name: "home",
			component: Home,
		},
		{
			path: "/login",
			name: "login",
			component: Login,
		},
		{
			path: "/register",
			name: "register",
			component: Register,
		},
		{
			path: "/items",
			name: "items",
			component: ItemsPage,
		},
		{
			path: "/items/new",
			name: "createItem",
			component: CreateItem,
		},
		// {
		// 	path: "/about",
		// 	name: "about",
		// 	// route level code-splitting
		// 	// this generates a separate chunk (about.[hash].js) for this route
		// 	// which is lazy-loaded when the route is visited.
		// 	component: () =>
		// 		import(/* webpackChunkName: "about" */ "./views/About.vue"),
		// },
	],
});

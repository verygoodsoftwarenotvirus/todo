<template>
	<div id="app">
		<section id="nav">
			<!--
			<b-tabs type="is-boxed">
				<router-link to="/">
					<b-tab-item label="Home" icon="photo_library"></b-tab-item>
				</router-link>
				<router-link to="/items">
					<b-tab-item label="Items" icon="photo_library"></b-tab-item>
				</router-link>
				<router-link to="/items/new">
					<b-tab-item
						label="New Item"
						icon="photo_library"
					></b-tab-item>
				</router-link>
				<router-link to="/register">
					<b-tab-item
						label="Register"
						icon="photo_library"
					></b-tab-item>
				</router-link>
				<router-link to="/login">
					<b-tab-item label="Login" icon="photo_library"></b-tab-item>
				</router-link>
			</b-tabs>
			-->

			<router-link to="/">Home</router-link> |
			<router-link to="/items">Items</router-link> |
			<router-link to="/items/new">New Item</router-link> |
			<router-link to="/register">Register</router-link> |
			<router-link to="/login">Login</router-link>
		</section>
		<router-view />
	</div>
</template>

<style>
#app {
	font-family: "Avenir", Helvetica, Arial, sans-serif;
	-webkit-font-smoothing: antialiased;
	-moz-osx-font-smoothing: grayscale;
	text-align: center;
	color: #666;
}
#nav {
	padding: 30px;
}

#nav a {
	font-weight: bold;
	color: #2c3e50;
}
</style>

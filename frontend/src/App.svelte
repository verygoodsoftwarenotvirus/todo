<script>
  import { Router, Link, Route } from "svelte-routing";
  import Home from "./pages/Home.svelte";
  import CreateItem from "./pages/CreateItem.svelte";
  import Items from "./pages/Items.svelte";
  import Login from "./pages/Login.svelte";
  import Register from "./pages/Register.svelte";

  export let url = "";
</script>

<!-- App.svelte -->

<Router {url}>
  <nav>
    <Link to="/">Home Page</Link>
    <Link to="items">Items</Link>
    <Link to="items/new">Create Item</Link>
    <Link to="login">Login</Link>
    <Link to="register">Register</Link>
  </nav>
  <div>
    <Route path="items" component={Items} />
    <Route path="items/new" component={CreateItem} />
    <Route path="login" component={Login} />
    <Route path="register" component={Register} />
    <Route path="/">
      <Home />
    </Route>
  </div>
</Router>

<script lang="typescript">
  import { Link } from "svelte-routing";

  // core components
  import IndexNavbar from "../components/Navbars/IndexNavbar.svelte";
  import Footer from "../components/Footers/MainFooter.svelte";

  export let location: Location;
</script>

<IndexNavbar />
<section class="header relative pt-16 items-center flex h-screen max-h-860-px">
  <div class="container mx-auto items-center flex flex-wrap">
    <div class="w-full md:w-8/12 lg:w-6/12 xl:w-6/12 px-4">
      <div class="pt-32 sm:pt-0">
        <h2 class="font-semibold text-4xl text-gray-700">
          This is the homepage.
        </h2>
        <p class="mt-4 text-lg leading-relaxed text-gray-600">
          Websites are cool and good to read.
        </p>
      </div>
    </div>
  </div>

</section>
<Footer />

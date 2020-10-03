<script lang="typescript">
  import { Link } from "svelte-routing";

  // core components
  import IndexNavbar from "components/Navbars/IndexNavbar.svelte";
  import Footer from "components/Footers/Footer.svelte";

  const patternVue: String = "/assets/img/pattern_svelte.png";
  const componentBtn: String = "/assets/img/component-btn.png";
  const componentProfileCard: String = "/assets/img/component-profile-card.png";
  const componentInfoCard: String = "/assets/img/component-info-card.png";
  const componentInfo2: String = "/assets/img/component-info-2.png";
  const componentMenu: String = "/assets/img/component-menu.png";
  const componentBtnPink: String = "/assets/img/component-btn-pink.png";
  const documentation: String = "/assets/img/documentation.png";
  const login: String = "/assets/img/login.jpg";
  const profile: String = "/assets/img/profile.jpg";
  const landing: String = "/assets/img/landing.jpg";

  export let location: Location;
</script>

<IndexNavbar />
<section class="header relative pt-16 items-center flex h-screen max-h-860-px">
  <div class="container mx-auto items-center flex flex-wrap">
    <div class="w-full md:w-8/12 lg:w-6/12 xl:w-6/12 px-4">
      <div class="pt-32 sm:pt-0">
        <h2 class="font-semibold text-4xl text-gray-700">
          A beautiful extension for Tailwind CSS.
        </h2>
        <p class="mt-4 text-lg leading-relaxed text-gray-600">
          Tailwind WebApp Svelte is Free and Open Source. It does not change or
          add any CSS to the already one from
          <a
            href="https://tailwindcss.com/?ref=creativetim"
            class="text-gray-700"
            target="_blank"
          >
            Tailwind CSS
          </a>
          . It features multiple HTML elements and it comes with dynamic
          components for ReactJS, Vue and Angular.
        </p>
        <div class="mt-12">
          <a
            href="https://www.creative-tim.com/learning-lab/tailwind/svelte/overview/webapp?ref=tws-index"
            target="_blank"
            class="get-started text-white font-bold px-6 py-4 rounded outline-none focus:outline-none mr-1 mb-1 bg-red-500 active:bg-red-600 uppercase text-sm shadow hover:shadow-lg ease-linear transition-all duration-150"
          >
            Get started
          </a>
          <a
            href="https://github.com/creativetimofficial/tailwind-webapp-svelte?ref=tws-index"
            class="github-star ml-1 text-white font-bold px-6 py-4 rounded outline-none focus:outline-none mr-1 mb-1 bg-gray-800 active:bg-gray-700 uppercase text-sm shadow hover:shadow-lg ease-linear transition-all duration-150"
            target="_blank"
          >
            Github Star
          </a>
        </div>
      </div>
    </div>
  </div>

  <img
    class="absolute top-0 b-auto right-0 pt-16 sm:w-6/12 -mt-48 sm:mt-0 w-10/12 max-h-860-px"
    src="{patternVue}"
    alt="..."
  />
</section>
<Footer />

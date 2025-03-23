# Samuel Boyden's Portfolio

This is the repository containing the source code for my portfolio hosted at
<https://stboyden.com>.

The code is written primarily using Golang as the language for the back-end
server, and using [Templ](https://templ.guide/) to generate server-side pages,
using [HTMX](https://htmx.org/) and [AlpineJS](https://alpinejs.dev/) to improve client-side interactivity and to emulate a
SPA feel. I also use [TailwindCSS](https://tailwindcss.com/) and [DaisyUI](https://daisyui.com/) to style the site, to make it look
simple yet consistent.

Additionally, I am implementing a custom blog system for this site, backed by
PostgresQL for storing posts.

## Running locally

If you would like to run a local copy, I recommend two commands that will make
your life easier. Both are *technically* optional, but I will recommend them
nonetheless.

1. `just` from [casey/just](https://github.com/casey/just) - Just is a task
   executor, similar to `make` but is more cross-platform which makes it more
   accessible for development on Windows. I personally develop on Linux, but the
   tasks should run fine on Windows. If you don't want to install this tool,
   that's fine, you'll just have to make sure you follow the commands in the
   `justfile` for the relevant commands to run for the task you want to do.
2. `docker` from <https://www.docker.com> - Docker is used to host a local
   development PostgresQL database for the blog system. If you already have a
   local PostgresQL instance that you want to use instead, make sure that you
   correctly set the `DB_URL` environment variable to compensate.

### Steps

0. **(Optional)** Create the PostgresQL instance from the `docker-compose.yml` file:

    ```bash
    docker compose up -d
    ```

1. Create a .env file with the following content:

    ```bash
    GITHUB_TOKEN=ghp_xxxxxxx # your GitHub PAT
    
    # the connection string to the database instance, see docker-compose.yml for the default connection string 
    DB_URL=xxxxx 
    ```

2. Run migrations on the database instance to get it to the correct state and
   generate types based on the updated schema:
  
    ```bash
    # this will run migrations first *then* generate the correct types
    just generate_db_types 
    ```

3. Build & run the project:

    ```bash
    just build
    just run
    ```

 > [!NOTE]
 >
 > Assuming no errors have been output, then the local instance will be
 > available at <http://localhost:8080>.

4. **(Optional)** If you want to have the site live-update with changes:

    ```bash
    just dev
    ```

 > [!NOTE]
 >
 > As above, your instance will be hosted at <http://localhost:8080>, but
 > after any changes you will have to manually refresh the page (with
 > <kbd>F5</kbd>) as I have not set up the project to work properly with
 > hot-reloading.

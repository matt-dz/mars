<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { resolve } from '$app/paths';
</script>

<div class="flex min-h-screen flex-col items-center justify-center px-4">
	<div class="text-center">
		<!-- Mars icon -->
		<div
			class="mx-auto mb-6 flex h-24 w-24 items-center justify-center rounded-full bg-gradient-to-br from-primary/20 via-accent/15 to-destructive/10"
		>
			<div
				class="flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-primary via-primary/80 to-destructive/60 shadow-lg shadow-primary/20"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 24 24"
					fill="currentColor"
					class="h-8 w-8 text-primary-foreground"
				>
					<circle cx="12" cy="12" r="10" />
				</svg>
			</div>
		</div>

		<!-- Error code -->
		<h1
			class="mb-2 bg-gradient-to-r from-primary to-destructive/80 bg-clip-text text-6xl font-bold tracking-tight text-transparent"
		>
			{page.status}
		</h1>

		<!-- Error message -->
		<h2 class="mb-4 text-2xl font-semibold text-foreground">
			{#if page.status === 404}
				Lost in space
			{:else if page.status === 500}
				Houston, we have a problem
			{:else if page.status === 403}
				Access denied
			{:else if page.status === 401}
				Authentication required
			{:else}
				Something went wrong
			{/if}
		</h2>

		<p class="mb-8 max-w-md text-muted-foreground">
			{#if page.status === 404}
				The page you're looking for has drifted into the void. It might have been moved or doesn't
				exist.
			{:else if page.status === 500}
				Our servers encountered an unexpected error. Our team has been notified and is working on
				it.
			{:else if page.status === 403}
				You don't have permission to access this resource. Please check your credentials.
			{:else if page.status === 401}
				Please log in to access this page.
			{:else}
				{page.error?.message || 'An unexpected error occurred. Please try again later.'}
			{/if}
		</p>

		<!-- Actions -->
		<div class="flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
			<Button href={resolve('/')} class="gap-2">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 20 20"
					fill="currentColor"
					class="h-4 w-4"
				>
					<path
						fill-rule="evenodd"
						d="M9.293 2.293a1 1 0 011.414 0l7 7A1 1 0 0117 11h-1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-3a1 1 0 00-1-1H9a1 1 0 00-1 1v3a1 1 0 01-1 1H5a1 1 0 01-1-1v-6H3a1 1 0 01-.707-1.707l7-7z"
						clip-rule="evenodd"
					/>
				</svg>
				Back to home
			</Button>
			{#if page.status === 401}
				<Button href={resolve('/login')} variant="outline" class="gap-2">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 20 20"
						fill="currentColor"
						class="h-4 w-4"
					>
						<path
							fill-rule="evenodd"
							d="M3 4.25A2.25 2.25 0 015.25 2h5.5A2.25 2.25 0 0113 4.25v2a.75.75 0 01-1.5 0v-2a.75.75 0 00-.75-.75h-5.5a.75.75 0 00-.75.75v11.5c0 .414.336.75.75.75h5.5a.75.75 0 00.75-.75v-2a.75.75 0 011.5 0v2A2.25 2.25 0 0110.75 18h-5.5A2.25 2.25 0 013 15.75V4.25z"
							clip-rule="evenodd"
						/>
						<path
							fill-rule="evenodd"
							d="M19 10a.75.75 0 00-.75-.75H8.704l1.048-.943a.75.75 0 10-1.004-1.114l-2.5 2.25a.75.75 0 000 1.114l2.5 2.25a.75.75 0 101.004-1.114l-1.048-.943h9.546A.75.75 0 0019 10z"
							clip-rule="evenodd"
						/>
					</svg>
					Log in
				</Button>
			{:else}
				<Button variant="outline" onclick={() => window.location.reload()} class="gap-2">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 20 20"
						fill="currentColor"
						class="h-4 w-4"
					>
						<path
							fill-rule="evenodd"
							d="M15.312 11.424a5.5 5.5 0 01-9.201 2.466l-.312-.311h2.433a.75.75 0 000-1.5H3.989a.75.75 0 00-.75.75v4.242a.75.75 0 001.5 0v-2.43l.31.31a7 7 0 0011.712-3.138.75.75 0 00-1.449-.39zm1.23-3.723a.75.75 0 00.219-.53V2.929a.75.75 0 00-1.5 0V5.36l-.31-.31A7 7 0 003.239 8.188a.75.75 0 101.448.389A5.5 5.5 0 0113.89 6.11l.311.31h-2.432a.75.75 0 000 1.5h4.243a.75.75 0 00.53-.219z"
							clip-rule="evenodd"
						/>
					</svg>
					Try again
				</Button>
			{/if}
		</div>
	</div>

	<!-- Decorative stars -->
	<div class="pointer-events-none fixed inset-0 overflow-hidden opacity-30 dark:opacity-50">
		<div class="absolute top-[20%] left-[10%] h-1 w-1 rounded-full bg-primary"></div>
		<div class="absolute top-[15%] left-[25%] h-0.5 w-0.5 rounded-full bg-muted-foreground"></div>
		<div class="absolute top-[25%] left-[80%] h-1.5 w-1.5 rounded-full bg-primary/60"></div>
		<div class="absolute top-[10%] left-[70%] h-0.5 w-0.5 rounded-full bg-muted-foreground"></div>
		<div class="absolute top-[70%] left-[15%] h-1 w-1 rounded-full bg-primary/40"></div>
		<div class="absolute top-[65%] left-[85%] h-0.5 w-0.5 rounded-full bg-muted-foreground"></div>
		<div class="absolute top-[80%] left-[45%] h-1 w-1 rounded-full bg-primary/50"></div>
		<div class="absolute top-[45%] left-[5%] h-0.5 w-0.5 rounded-full bg-muted-foreground"></div>
		<div class="absolute top-[40%] left-[92%] h-1 w-1 rounded-full bg-primary/30"></div>
	</div>
</div>

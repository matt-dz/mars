<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	let email = $state('');
	let password = $state('');
	let isLoading = $state(false);
	let error = $state('');

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		isLoading = true;

		try {
			const response = await fetch('/api/login', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ email, password })
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.message || 'Login failed');
			}

			goto(resolve('/'));
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
			console.error(error);
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="flex min-h-screen flex-col items-center justify-center px-4">
	<div class="mb-8 text-center">
		<div
			class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-primary via-primary/80 to-destructive/60 shadow-lg shadow-primary/20"
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
		<h1
			class="bg-gradient-to-r from-primary to-destructive/80 bg-clip-text text-4xl font-bold tracking-tight text-transparent"
		>
			mars
		</h1>
		<p class="mt-2 text-muted-foreground">Your personal music playlist generator</p>
	</div>

	<Card.Root class="w-full max-w-sm overflow-hidden shadow-lg">
		<div class="h-1.5 bg-gradient-to-r from-primary via-primary/70 to-destructive/50"></div>
		<Card.Header class="space-y-1 pt-6">
			<Card.Title class="text-2xl font-bold">Welcome back</Card.Title>
			<Card.Description>Enter your credentials to access your account</Card.Description>
		</Card.Header>
		<Card.Content>
			<form onsubmit={handleSubmit} class="space-y-4">
				{#if error}
					<div
						class="flex items-center gap-2 rounded-lg bg-destructive/10 p-3 text-sm text-destructive"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 20 20"
							fill="currentColor"
							class="h-5 w-5 shrink-0"
						>
							<path
								fill-rule="evenodd"
								d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-5a.75.75 0 01.75.75v4.5a.75.75 0 01-1.5 0v-4.5A.75.75 0 0110 5zm0 10a1 1 0 100-2 1 1 0 000 2z"
								clip-rule="evenodd"
							/>
						</svg>
						{error}
					</div>
				{/if}

				<div class="space-y-2">
					<Label for="email">Email</Label>
					<Input
						id="email"
						type="email"
						placeholder="name@example.com"
						bind:value={email}
						required
						disabled={isLoading}
						class="h-11"
					/>
				</div>

				<div class="space-y-2">
					<Label for="password">Password</Label>
					<Input
						id="password"
						type="password"
						bind:value={password}
						required
						disabled={isLoading}
						class="h-11"
					/>
				</div>

				<Button type="submit" class="h-11 w-full gap-2" disabled={isLoading}>
					{#if isLoading}
						<svg
							class="h-4 w-4 animate-spin"
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
						>
							<circle
								class="opacity-25"
								cx="12"
								cy="12"
								r="10"
								stroke="currentColor"
								stroke-width="4"
							></circle>
							<path
								class="opacity-75"
								fill="currentColor"
								d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
							></path>
						</svg>
						Logging in...
					{:else}
						Sign in
					{/if}
				</Button>
			</form>
		</Card.Content>
	</Card.Root>
</div>

<script lang="ts">
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Avatar from '$lib/components/ui/avatar';
	import { goto } from '$app/navigation';
	import type { User } from '$lib/api/types';
	import { resolve } from '$app/paths';

	let { user }: { user: User } = $props();

	function navigateToIntegrations() {
		goto(resolve('/integrations'));
	}

	async function handleLogout() {
		await fetch('/api/logout', { method: 'POST' });
		goto(resolve('/login'));
	}
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger class="rounded-full ring-ring outline-none focus-visible:ring-2">
		<Avatar.Root class="h-9 w-9">
			<Avatar.Fallback class="bg-primary text-primary-foreground">
				{user.email[0].toUpperCase()}
			</Avatar.Fallback>
		</Avatar.Root>
	</DropdownMenu.Trigger>
	<DropdownMenu.Content align="end" class="w-48">
		<DropdownMenu.Label class="font-normal">
			<p class="truncate text-sm text-muted-foreground">{user.email}</p>
		</DropdownMenu.Label>
		<DropdownMenu.Separator />
		<DropdownMenu.Item onSelect={navigateToIntegrations}>Integrations</DropdownMenu.Item>
		<DropdownMenu.Separator />
		<DropdownMenu.Item onSelect={handleLogout}>Logout</DropdownMenu.Item>
	</DropdownMenu.Content>
</DropdownMenu.Root>

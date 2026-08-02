# 03 — Hide inventory, equipment and vendor

**What to build:** the player sees no inventory, no equipment screen and no shop, and has
no way to reach them — while all of that code remains present and compiling.

This is deliberate mothballing, not deletion. Inventory, item, equipment, pickup, drop,
item-move and item-hook code stays, along with the shop interface, vendor definitions and
the vendor branches of the talk action. Keybindings and menu entries are what go.

The distinction matters: this code represents real work that will be wanted back, and
rewriting it later would be worse than carrying it unreferenced now.

**Blocked by:** 02 — Demolish the RPG systems.

**Status:** ready-for-agent

- [ ] No inventory, equipment or shop interface is reachable by any input
- [ ] Talking to a Townsperson opens no vending options
- [ ] Item, inventory, equipment and vendor code still compiles
- [ ] Automatic item pickup no longer runs
- [ ] Nothing in the interface refers to items, gold or equipment

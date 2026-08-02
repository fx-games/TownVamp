# Town Design Handover
## Vampire Town Proof of Concept

This document describes the initial authored content for the proof-of-concept town simulation.

The goal is **not** to tell a story through dialogue, but to create a small town whose inhabitants have understandable routines, visible responsibilities, and private lives that can be observed.

The player should learn almost everything by watching people move through the world.

---

# Design Principles

Every Townsperson consists of three layers.

## Routine

The things everyone does.

Examples:

- Sleep
- Eat breakfast
- Eat dinner
- Spend evening at home

The routine is public and predictable.

---

## Obligations

Public duties that keep the town functioning.

These are what make the dependency graph.

Examples:

- Deliver vegetables
- Grind grain
- Repair tools
- Hold church service

Obligations should create movement between locations and interactions with other Townspeople.

---

## Secret

An optional repeated activity.

Secrets should:

- explain unusual movement
- usually happen in private
- be observable
- never require exposition
- often be morally ambiguous rather than obviously evil

Secrets should mostly consist of ordinary actions:

- walking
- carrying an item
- interacting with scenery
- standing still
- talking to oneself

---

# Townspeople

---

# Elias
Role: Farmer

## Routine

- Sleep at farmhouse
- Breakfast
- Work vegetable field
- Lunch under tree
- Continue working
- Dinner
- Sleep

## Obligations

- Harvest vegetables
- Deliver vegetables to the inn
- Occasionally bring grain to the mill

## Secret

After sunset he visits the grave of his deceased daughter.

Visible behaviour:

- Carry flowers
- Walk to cemetery
- Stand silently
- Return home

---

# Martin
Role: Miller

## Routine

- Sleep
- Open mill
- Grind grain
- Lunch
- Maintain waterwheel
- Close mill
- Sleep

## Obligations

- Grind grain delivered by the farmer
- Deliver flour to the inn

## Secret

Sometimes sleepwalks to the river during the night.

Visible behaviour:

- Leave house
- Walk to river
- Stand still
- Return

---

# Greta
Role: Innkeeper

## Routine

- Prepare breakfast
- Open inn
- Serve meals
- Clean common room
- Close inn
- Sleep

## Obligations

- Feed townspeople
- Feed travellers
- Buy vegetables from farmer
- Buy flour from miller

## Secret

Waters down the ale.

Visible behaviour:

- Carry bucket
- Enter cellar
- Interact with ale barrel
- Leave

---

# Otto
Role: Blacksmith

## Routine

- Open forge
- Smith tools
- Lunch
- Continue working
- Close forge
- Sleep

## Obligations

- Repair farming tools
- Repair carpenter's tools

## Secret

Collects discarded iron scraps after customers leave.

Visible behaviour:

- Wait until alone
- Pick up iron scraps
- Carry them into storage

---

# Marta
Role: Carpenter

## Routine

- Breakfast
- Workshop
- Repair jobs around town
- Lunch
- Workshop
- Sleep

## Obligations

- Repair mill
- Repair inn furniture
- Repair houses

## Secret

Carves wooden toys every night.

Visible behaviour:

- Carry wood
- Sit at workbench
- Carve
- Hide toy in drawer

---

# Father Aldren
Role: Priest

## Routine

- Morning prayer
- Church duties
- Visit villagers
- Evening prayer
- Sleep

## Obligations

- Hold services
- Conduct funerals
- Visit the sick
- Bless the fields

## Secret

Has lost his faith and secretly rehearses every sermon from written notes.

Visible behaviour:

- Enter study
- Read notes
- Pace
- Repeat quietly

---

# Dependency Graph

Farmer
→ supplies vegetables to Inn

Farmer
→ supplies grain to Miller

Miller
→ supplies flour to Inn

Blacksmith
→ repairs Farmer's tools

Blacksmith
→ repairs Carpenter's tools

Carpenter
→ repairs Mill

Carpenter
→ repairs Inn

Priest
→ interacts with everyone

This graph is intentionally small and easy to understand.

---

# Required Locations

## Public

- Town Square
- Main Road
- Inn
- Blacksmith
- Carpenter Workshop
- Church
- Cemetery
- Farmhouse
- Vegetable Field
- Grain Field
- Mill
- River
- Bridge

---

## Semi-private

- Inn Cellar
- Mill Interior
- Smith Storage
- Carpenter Workshop Interior
- Priest Study
- Barn

---

## Private

- Farmer Bedroom
- Miller Bedroom
- Innkeeper Bedroom
- Blacksmith Bedroom
- Carpenter Bedroom
- Priest Bedroom

---

## Secret Locations

- Child's Grave
- Quiet Riverbank

These exist solely to support secrets.

---

# Required Scenery

Farm

- Crop rows
- Water bucket
- Wheelbarrow
- Hay bale
- Fence
- Tree

Mill

- Mill wheel
- Grain chute
- Grain sacks
- Flour sacks
- Storage shelves

Inn

- Tables
- Chairs
- Fireplace
- Counter
- Barrels
- Kitchen

Blacksmith

- Forge
- Anvil
- Grindstone
- Tool rack
- Coal pile

Carpenter

- Workbench
- Sawhorse
- Lumber pile
- Shelves

Church

- Altar
- Benches
- Lectern
- Candles
- Bell rope

Cemetery

- Graves
- Flowers
- Fence
- Gate

River

- Dock
- Bushes
- Waterwheel
- Footbridge

---

# Portable Items

Food

- Vegetables
- Grain Sack
- Flour Sack
- Bread
- Ale Barrel

Smith

- Hammer
- Tongs
- Horseshoe
- Hoe
- Axe Head
- Iron Scrap

Carpenter

- Board
- Log
- Saw
- Hammer
- Wooden Toy

Church

- Prayer Book
- Prayer Notes
- Flowers
- Candle

Miscellaneous

- Bucket
- Basket
- Mug
- Coin Purse
- Lantern

---

# Observable Actions

The simulation should be readable from simple repeated actions.

Farmer

- Harvest
- Carry basket
- Deliver basket
- Place flowers

Miller

- Carry sack
- Operate mill
- Carry flour

Innkeeper

- Carry barrel
- Serve food
- Clean table
- Pour water into barrel

Blacksmith

- Hammer
- Sharpen
- Carry tools
- Collect scraps

Carpenter

- Saw wood
- Hammer
- Carry boards
- Carve toy

Priest

- Pray
- Read
- Ring bell
- Place flowers
- Rehearse sermon

No bespoke animations are required.

Movement, interaction, waiting and speech are sufficient.

---

# Conversations

Dialogue should be extremely short.

Its purpose is to reinforce relationships rather than tell stories.

## Farmer ↔ Innkeeper

"Morning."

"Fresh vegetables?"

"As always."

---

## Farmer ↔ Miller

"More grain."

"I'll have flour tomorrow."

---

## Miller ↔ Innkeeper

"Here's today's flour."

"Leave it in the pantry."

---

## Carpenter ↔ Miller

"The wheel is squeaking again."

"I'll come tomorrow."

---

## Carpenter ↔ Innkeeper

"Another broken chair?"

"The travellers never sit gently."

---

## Blacksmith ↔ Farmer

"Need this repaired."

"I'll have it done."

---

## Blacksmith ↔ Carpenter

"Broke another hammer?"

"It happens."

---

## Priest ↔ Farmer

"The harvest looks promising."

"We've been fortunate."

---

## Priest ↔ Innkeeper

"Busy today?"

"Busy enough."

---

## Generic greetings

"Good morning."

"Good evening."

"See you tomorrow."

---

# Secret Dialogue

Secrets should mostly happen without dialogue.

If dialogue exists, it should be brief.

Farmer

"I still miss you."

---

Innkeeper

"No one will notice."

---

Carpenter

"Just one more."

---

Priest

"I hope they believe me."

---

# Content Goals

The player should be able to infer:

- where people usually are
- who depends on whom
- what everyone's job is
- which actions are ordinary
- which actions are unusual

without ever being explicitly told.

The town should feel alive because people visibly fulfil obligations to one another.

Secrets exist only to create interesting movement patterns and relationships, not to drive quests or unlock scripted events.

Everything should be expressible using only:

- movement
- speech
- carrying items
- interacting with scenery
- waiting

No custom gameplay mechanics should be required for any authored behaviour.

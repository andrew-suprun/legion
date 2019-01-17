# Philosophy
- Experimenting with variation of events sourcing I call (tongue in cheek) as "command sourcing".
  As advocates of event sourcing insist that every stored event record reflects actual business 
  relevant event, the best way to do it would be to store the command that caused the event and 
  reference all events to origination commands.
- Ironically, it pushed events into second class citizens.

# Archtecture
- major architectural parts
  - commands
  - queries
  - events
  - entities AKA aggregates
  - projections
    - used mostly by queries
  - reactors
    - used mostly for 3rd party integration

- 2 databases
  - events
    - events
    - commands
  - entities
    - entity per entity type
    - snapshot per entity type
      - potentially large aggregates may be split into sub-collections

- versioning
  - for changes that are both backward and forward compatible no versioning is needed
  - for incompatible changes create new commands/events

# Framework
- components are structured by business function and not by structural properties
  - for example all user related command handlers, reactors, projections are part 
    of the users component

- event has:
  - event id                   (key)
  - sequence counter           (indexed)
  - entity id                  (indexed)
  - command id                 (indexed)
  - timestamp                  (indexed)
  - info

- entity has:  
  - entity id                  (key)
  - current value of sequence counter

- snapshot has:
  - entity id                  (key)
  - latest value of sequence counter
  - aggregated entity

- track incoming and outgoing payments in separate collections
  - payable
  - receivable

- tester can run in 2 modes
  - in-memory
  - using mongo

# Design
- three separate product categories
  - spending accounts
  - saving accounts
  - insurance

# Strategies
- backup strategy
  - incremental backup only need to store command and events since last backup

- strategy to trim database
  - select cutout timestamp
  - snapshot all entities at cutout time in new database
  - populate new database with events past cutout time
  - switch to the new database
  - archive old database

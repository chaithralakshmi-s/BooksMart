# **Team Starburst Hackathon Project Minutes Week 3**

## Meeting Date and Time

**Date:** 14th April 2018</br>
**Location:** Engineering Building</br>
**Time:** 11.45AM - 12:45PM</br>
**Attendees:**

1. Chaithra
2. Luis
3. Mithun
4. Sowmya
5. Radhika

## Action Items from Last Week

1. Created Slack group for project.
2. Created Functional Diagram.

## Key Discussion Points

1. Alloted API and Database to members
2. Technology to be used for API creation
3. Database to be used Mongo, Redis or Riak


## Decisions

1. Finalized on who will be developing which API


## Action Items

1. Finalize the design and architecture.
2. Decide on how the super user can share his shopping cart? - how to generate the link, will it be a random code?
3. Decide which databse to implemented based on personal project
4. Design the database.
5. Discuss on how to co-ordinate the APIs.


## Questions

1. How to sync the common shopping cart among users?
2. Should the all users be able to edit/delete products that were added by others?
3. Should only Redis be used, Or Redis used as a in-mem cache along with Mongo/Riak
4. Should the webapp have login/signup and invitation to add user to the cart, or should the invitiation be just via a link.
5. Can we use AWS Lambda or other serverless functions for the backend API call, and then link Lambda with DynamoDB or other examples.
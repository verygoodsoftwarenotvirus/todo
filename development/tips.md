# Tips for working in this codebase

## FAQ

**Q: My integration test is failing, but I don't think it should be, what gives?**

**A:** In my experience, the answer is usually:
1. Routing: the router isn't sending your route to where you think it is. Check other routes in the same sub-router with the same method.
1. Client: integration tests should use a client they create for their own test, but most don't. 
    - Are you using a client that has access to the resource you're trying to test?
    - Are you testing a new route, and if so, does the client use the correct route?
##


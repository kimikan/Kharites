





<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
  </head>

  <body>
    <h3>Description</h3>
   In 2016, rewrite a old c++ service, with go.   And I extract the main framework to this open source project
 
   And finally I remove most of the related bussiness logic.  and rename it with Kharites.
   and opensource it to github

   <h4><p>
   The architecture guidelines. <br/>
   1. All of the IO operations to a seperate dev will be scheduled in a single same goroutine.<br/>
   2. The requested size of block splitted(64k).<br/>
   3. The disk data can be cached in memory with freecache.<br/>
   4. Zero copy<br/>
   5. Preallocated memory.<br/>
   
   </p></h4>
<h3>Usage</h3>

Modify the config to your server's disk.
the start the server.
The current logic is read raw dev with pread(). And with a block of "64k" equals 128 sectors.

you can modify it whatever you want to fit your bussiness.

<h3>Dependencies</h3>
https://github.com/coocood/freecache

<h3>Performance</h3>
The IO is very good.

Due to the goroutine schedule.  The cpu usage is a little higher than c++ epoll mechanism.
but it's graceful.

  </body>
</html>


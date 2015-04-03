# jettison

*destroys all ephemeral containers on a garden server*

![Fuel Dump](https://farm4.staticflickr.com/3245/2412171473_21a703d392_z_d.jpg?zz=1)

by [paperghost](https://www.flickr.com/photos/paperghost/2412171473)

## about

Shutting a Concourse worker down requires the containers to be destroyed before removing the old versions of the RootFSs so that the files are not in use when BOSH tries to remove them. The only containers that are likely to last long enough to be in this state (over two BOSH deploys) are *check* containers.

Luckily, *check* containers can be destroyed at any time and spun up again on another machine to resume. In Concourse this concept is referred to as *ephemeral*. When Concourse creates an *ephemeral* container it is marked with a Garden property. *Jettison* finds containers with this property and destroys them.

## usage

To destroy all ephemeral containers on a Garden server, just run:

```bash
$ jettison -gardenAddr 10.0.0.100:7777
```

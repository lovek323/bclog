# Bigcommerce Syslog Formatter
Here, I fixed it.

## Configuration
Copy `config.json` to `~/.config/bclog/config.json` and set `PrimaryKeyFile` to
`/Users/your.username/.vagrant.d/insecure_private_key`.

You can modify the types of messages you ignore.

## Message format

Messages are displayed in the following format:

```
[id]  timestamp  source/category  description
```

## Viewing detailed information for an individual message

Just type the id of a message to view more information:

```
> 6701

---------- PHP LOG EVENT ----------
SyslogTime:  0000-07-09 02:25:33
LogLevel:    Warning
Content:     rename(/tmp/__CG__BigcommerceCatalogDataModelBrand.php.53bc8bfd186be2.93041831,/tmp/__CG__BigcommerceCatalogDataModelBrand.php): Operation not permitted
File:        /opt/bigcommerce_app/vagrant_code/vendor/doctrine/common/lib/Doctrine/Common/Proxy/ProxyGenerator.php
Line:        305

Stack trace
0.                                                                                                                                                                                             0
1.   {main}                                                                /opt/bigcommerce_app/vagrant_code/index.php                                                                         0
2.   Interspire_RequestDispatcher->dispatch                                /opt/bigcommerce_app/vagrant_code/index.php                                                                         60
3.   Interspire_RequestDispatcher->followRoute                             /opt/bigcommerce_app/vagrant_code/lib/Interspire/RequestDispatcher.php                                              196
4.   Interspire_RequestRoute->processFollow                                /opt/bigcommerce_app/vagrant_code/lib/Interspire/RequestDispatcher.php                                              117
5.   Store\RequestRoute\ApiV3Action->_follow                               /opt/bigcommerce_app/vagrant_code/lib/Interspire/RequestRoute.php                                                   30
6.   Api\V3\ResourceController->handleRequest                              /opt/bigcommerce_app/vagrant_code/lib/Store/RequestRoute/ApiV3Action.php                                            45
7.   Api\V3\ResourceController->handleAction                               /opt/bigcommerce_app/vagrant_code/app/controllers/Api/V3/ResourceController.php                                     260
8.   call_user_func                                                        /opt/bigcommerce_app/vagrant_code/app/controllers/Api/V3/ResourceController.php                                     293
9.   Api\V3\CatalogResourceController->indexAction                         /opt/bigcommerce_app/vagrant_code/app/controllers/Api/V3/ResourceController.php                                     293
10.  Bigcommerce\Catalog\Service\Impl\BaseService->findAllPaged            /opt/bigcommerce_app/vagrant_code/app/controllers/Api/V3/CatalogResourceController.php                              67
11.  Doctrine\ORM\EntityRepository->findBy                                 /opt/bigcommerce_app/vagrant_code/vendor/bigcommerce/catalog-service/src/Service/Impl/BaseService.php               138
12.  Doctrine\ORM\Persisters\BasicEntityPersister->loadAll                 /opt/bigcommerce_app/vagrant_code/vendor/doctrine/orm/lib/Doctrine/ORM/EntityRepository.php                         181
13.  Doctrine\ORM\Internal\Hydration\AbstractHydrator->hydrateAll          /opt/bigcommerce_app/vagrant_code/vendor/doctrine/orm/lib/Doctrine/ORM/Persisters/BasicEntityPersister.php          934
14.  Doctrine\ORM\Internal\Hydration\SimpleObjectHydrator->hydrateAllData  /opt/bigcommerce_app/vagrant_code/vendor/doctrine/orm/lib/Doctrine/ORM/Internal/Hydration/AbstractHydrator.php      140
15.  Doctrine\ORM\Internal\Hydration\SimpleObjectHydrator->hydrateRowData  /opt/bigcommerce_app/vagrant_code/vendor/doctrine/orm/lib/Doctrine/ORM/Internal/Hydration/SimpleObjectHydrator.php  48
16.  Doctrine\ORM\UnitOfWork->createEntity                                 /opt/bigcommerce_app/vagrant_code/vendor/doctrine/orm/lib/Doctrine/ORM/Internal/Hydration/SimpleObjectHydrator.php  132
17.  Doctrine\Common\Proxy\AbstractProxyFactory->getProxy                  /opt/bigcommerce_app/vagrant_code/vendor/doctrine/orm/lib/Doctrine/ORM/UnitOfWork.php                               2663
18.  Doctrine\Common\Proxy\AbstractProxyFactory->getProxyDefinition        /opt/bigcommerce_app/vagrant_code/vendor/doctrine/common/lib/Doctrine/Common/Proxy/AbstractProxyFactory.php         119
19.  Doctrine\Common\Proxy\ProxyGenerator->generateProxyClass              /opt/bigcommerce_app/vagrant_code/vendor/doctrine/common/lib/Doctrine/Common/Proxy/AbstractProxyFactory.php         218
20.  rename                                                                /opt/bigcommerce_app/vagrant_code/vendor/doctrine/common/lib/Doctrine/Common/Proxy/ProxyGenerator.php               305
```

## Commands

Commands and some arguments can be tab completed. The following commands are
available: `clear` (clears the screen), `reload` (reloads your config file),
`show` (shows details for a particular category of message), `quit` (quits the
program) and `summary` (shows a summary of all events over a timeframe).

For online help, type `help`.

### Showing a summary

#### Default summary

Press `<RETURN>` on an empty line to see a summary of the events since you last
pushed `<RETURN>`.

```
SUMMARY (LAST 1m50.683039541s)
nginx-access-200  2 event(s)  Last 0 ago
```

#### Last 24 hours

The `summary` command will show you a summary of errors for the past 24 hours.

```
SUMMARY (LAST 24h0m0s)
php-Warning                333 event(s)   Last 12m3s ago
nginx-access-302           34 event(s)    Last 40m52s ago
nginx-access-404           10 event(s)    Last 40m50s ago
generic-rsyslogd           2 event(s)     Last 20h13m40s ago
php-Fatal error            13 event(s)    Last 13m9s ago
php-stack-trace            4479 event(s)  Last 12m3s ago
nginx-access-204           6 event(s)     Last 36m46s ago
php-Catchable fatal error  1 event(s)     Last 12m19s ago
process                    322 event(s)   Last 8m41s ago
nginx-access-200           1535 event(s)  Last 0 ago
php-Notice                 5 event(s)     Last 41m31s ago
php-Parse error            2 event(s)     Last 19h0m45s ago
php-SQL Error              75 event(s)    Last 1h27m3s ago
nginx-access-401           62 event(s)    Last 38m45s ago
nginx-access-304           8 event(s)     Last 36m54s ago
generic-bcdeploy           2 event(s)     Last 38m27s ago
generic-bigcommerce        30 event(s)    Last 38m27s ago
```

#### Custom timeframe

The `summary` command takes an optional argument, a time specifier, e.g.,
`summary 3h` for a summary of the last three hours. Valid time units (accepted
by `time.ParseDuration()` are 'ns' (nanoseconds), 'us' (microseconds), 'ms'
(milliseconds), 's' (seconds), 'm' (minutes) and h (hours).

### Showing events of a particular category

Type `show <type> <duration>` (where type is the event type, e.g., `php`) to
show all events of this type for a particular timeframe.

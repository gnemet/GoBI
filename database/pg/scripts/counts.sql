select 'vir_vir01' as name , count(1) as count from vir_vir01 union all
select 'vir_vir02' as name , count(1) from vir_vir02 union all
select 'vir_vir03' as name , count(1) from vir_vir03 union all
select 'vir_vir04' as name , count(1) from vir_vir04 union all
select 'vir_vir05' as name , count(1) from vir_vir05 union all
select 'vir_vir06' as name , count(1) from vir_vir06 union all
select 'vir_vir08' as name , count(1) from vir_vir08 union all
select 'vir_vir09' as name , count(1) from vir_vir09 union all
select 'vir_vir10' as name , count(1) from vir_vir10 union all
select 'vir_vir11' as name , count(1) from vir_vir11 ;


select * from vir_param ;


select  *
from vir_vir10 as a
left join vir_vir10_d as d on a.id = d.id
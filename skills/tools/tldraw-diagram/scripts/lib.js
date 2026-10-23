editor.deleteShapes([...editor.getCurrentPageShapeIds()]);
const rt=(lines)=>({type:'doc',content:[].concat(lines).map(l=>({type:'paragraph',content:l?[{type:'text',text:l}]:[]}))});
const id=(n)=>'shape:'+n;
const S=[];
const text=(n,x,y,t,o={})=>S.push({id:id(n),type:'text',x,y,props:{richText:rt(t),color:o.color||'black',size:o.size||'s',font:o.font||'draw',autoSize:true}});
const dot=(n,x,y,color,r=14)=>S.push({id:id(n),type:'geo',x,y,props:{geo:'ellipse',w:r,h:r,color,fill:'fill',dash:'solid',size:'s'}});
const line=(n,x,y,w)=>S.push({id:id(n),type:'line',x,y,props:{color:'grey',dash:'solid',size:'s',points:{a1:{id:'a1',index:'a1',x:0,y:0},a2:{id:'a2',index:'a2',x:w,y:0}}}});
// card: white box with a pill tag (border + text in `color`, light fill)
const card=(n,x,y,w,h,tag,color,lines)=>{
  S.push({id:id(n),type:'geo',x,y,props:{geo:'rectangle',w,h,color:'black',fill:'semi',dash:'draw',size:'m',font:'draw',align:'start',verticalAlign:'start',labelColor:'black',richText:rt(['',...lines])}});
  const tw=tag.length*10+52;
  S.push({id:id(n+'_tag'),type:'geo',x:x+14,y:y-17,props:{geo:'oval',w:tw,h:34,color,fill:'solid',dash:'solid',size:'s'}});
  // separate text shape: geo labels ignore scale and need ~50px height (growY lint)
  S.push({id:id(n+'_tl'),type:'text',x:x+20,y:y-11,props:{w:tw-12,richText:rt('● '+tag),color,size:'s',font:'draw',autoSize:false,textAlign:'middle',scale:0.9}});
};
const band=(n,y,label,color,W)=>{line(n+'_l',40,y,W);dot(n+'_d',48,y+30,color,16);text(n+'_t',74,y+18,label,{size:'m'});};
const A=[];
const arrow=(a,b,color,o={})=>A.push({a,b,color,...o});
const finish=async(title,sub)=>{
  text('title',40,20,title,{size:'xl'}); text('sub',42,80,sub,{size:'s',color:'grey'});
  editor.createShapes(S);
  A.forEach((r,i)=>{const aid=id('arr'+i);
    editor.createShape({id:aid,type:'arrow',x:0,y:0,props:{color:r.color,dash:r.dash||'draw',size:'m',bend:r.bend||0,kind:r.elbow?'elbow':'arc',font:'draw',richText:rt(r.label||''),start:{x:0,y:0},end:{x:1,y:1}}});
    const bind=(t,s,p)=>({fromId:aid,toId:id(s),type:'arrow',props:{terminal:t,normalizedAnchor:p||{x:.5,y:.5},isExact:false,isPrecise:!!p}});
    editor.createBindings([bind('start',r.a,r.from),bind('end',r.b,r.to)]);});
  editor.updateInstanceState({isGridMode:false}); editor.user.updateUserPreferences({colorScheme:'light'});
  editor.zoomToFit(); await helpers.saveDoc(); return {n:editor.getCurrentPageShapeIds().size, lints:helpers.getLints()};
};

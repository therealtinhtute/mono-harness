const ids=[...editor.getCurrentPageShapeIds()];
const {blob} = await editor.toImage(ids,{format:'png',background:true,padding:32,scale:1,darkMode:false});
const buf = new Uint8Array(await blob.arrayBuffer()); let s=''; for (let i=0;i<buf.length;i+=0x8000) s+=String.fromCharCode(...buf.subarray(i,i+0x8000));
return btoa(s);

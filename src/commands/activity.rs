use log::*;
use serenity::{
    prelude::*,
    model::{
        channel::{Message},
        gateway::{Activity},
    },
    framework::standard::{
        Args, CommandResult,
        macros::*,
    },
};

#[command]
#[bucket = "activity"]
#[aliases("set")]
async fn play(ctx: &Context, msg: &Message, args: Args) -> CommandResult {
    let mut name = args.message().to_string();

    for user in &msg.mentions {
        name = name.replace(user.to_string().as_str(), user.name.as_str());
    }
    
    
    ctx.set_activity(Activity::playing(&name)).await;

    Ok(())
}
